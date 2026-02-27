package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/xiehqing/common/datasource"
	"github.com/xiehqing/common/pkg/crypto"
	"github.com/xiehqing/common/pkg/httpx"
	"github.com/xiehqing/common/pkg/logs"
	"github.com/xiehqing/common/pkg/util"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	Url           string
	client        *httpx.Client
	config        *Config
	customHeaders map[string]string
}

// NewClient 创建一个新的Client实例
func NewClient(config *Config) (*Client, error) {
	customHeaders := map[string]string{
		"Connection":   "keep-alive",
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	for k, v := range config.Headers {
		customHeaders[k] = v
	}
	if config.Username != "" && config.Password != "" {
		authToken, err := crypto.Base64Crypto.Encrypt(fmt.Sprintf("%s:%s", config.Username, config.Password))
		if err != nil {
			logs.Errorf("认证信息编码失败：%v, 请检查配置", err)
		} else {
			customHeaders["Authorization"] = fmt.Sprintf("Basic %s", authToken)
		}
	}
	c := &Client{
		config:        config,
		Url:           config.Url,
		customHeaders: customHeaders,
	}
	err := c.Init()
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (cli *Client) Init() error {
	config := cli.config
	client := httpx.NewTransportClient(config.Url,
		&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(config.DailTimeout) * time.Millisecond,
				KeepAlive: 30 * time.Millisecond,
			}).DialContext,
			ResponseHeaderTimeout: time.Duration(config.Timeout) * time.Millisecond,
			MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		}, time.Duration(config.Timeout)*time.Millisecond)
	cli.client = client
	err := cli.CheckConnectivity()
	if err != nil {
		return err
	}
	return nil
}

// Equal 判断是否相同数据源
func (cli *Client) Equal(other datasource.DataSource) bool {
	otherCli, ok := other.(*Client)
	if !ok {
		logs.Errorf("数据源类型不匹配")
		return false
	}
	for k, v := range cli.customHeaders {
		if otherCli.customHeaders[k] != v {
			return false
		}
	}
	for k, v := range otherCli.customHeaders {
		if cli.customHeaders[k] != v {
			return false
		}
	}
	return cli.Url == otherCli.Url
}

// FastCheckConnectivity 快速测试连通性
func FastCheckConnectivity(url, username, password string) error {
	_, err := NewClient(&Config{Url: url, Username: username, Password: password, Timeout: 1000})
	if err != nil {
		return err
	}
	return nil
}

// CheckConnectivity 测试连通性
func (cli *Client) CheckConnectivity() error {
	uri := fmt.Sprintf("%s/flags", cli.Url)
	resp, err := cli.client.Client.Get(uri)
	if err != nil {
		return errors.WithMessagef(err, "测试连通性失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("连接失败, 状态码：%d", resp.StatusCode)
	}
	logs.Debugf("连接成功, 状态码：%d", resp.StatusCode)
	return nil
}

// QueryData 查询数据
func (cli *Client) QueryData(ctx context.Context, query interface{}) ([]*datasource.MetricPoint, error) {
	if query == nil {
		return nil, errors.Errorf("query is nil")
	}
	queryParam, err := util.Convert[QueryParam](query)
	if err != nil {
		return nil, errors.WithMessagef(err, "[prometheus] query param convert error")
	}
	err = queryParam.Validate()
	if err != nil {
		return nil, errors.WithMessagef(err, "[prometheus] query param validate error")
	}
	if queryParam.Time == nil {
		return nil, errors.Errorf("[prometheus] 查询时间点不得为空.")
	}
	metricPoints, err := cli.Query(queryParam.Query, *queryParam.Time, queryParam.Step)
	if err != nil {
		return nil, errors.WithMessagef(err, "[prometheus] query error")
	}
	return metricPoints, nil
}

// handlePrometheusResponse 处理Prometheus响应
func (cli *Client) handlePrometheusResponse(resp *http.Response, body []byte) (model.Value, error) {
	statusCode := resp.StatusCode
	if !apiError(statusCode) && statusCode/100 != 2 {
		return nil, errors.Errorf("请求失败，状态码：%d", statusCode)
	}
	var response Response
	if statusCode != http.StatusNoContent {
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, errors.WithMessagef(err, "解析响应失败")
		}
	}
	if response.Status != "success" {
		return nil, errors.Errorf("请求失败:%s", response.Error)
	}
	var qr QueryResult
	err := json.Unmarshal(response.Data, &qr)
	if err != nil {
		return nil, errors.WithMessagef(err, "解析响应失败")
	}
	return qr.v, nil
}

// Query 查询数据
func (cli *Client) Query(query string, timestamp time.Time, step string) ([]*datasource.MetricPoint, error) {
	uri := fmt.Sprintf("%s/query", apiPrefix)
	opt := httpx.NewRequestOption(
		httpx.WithQueryParam("query", query),
		httpx.WithQueryParam("time", formatTime(timestamp)),
		httpx.WithQueryParam("step", step),
		httpx.WithHeaders(cli.customHeaders),
		httpx.WithMethod(http.MethodGet),
		httpx.WithPath(uri),
	)
	resp, err := cli.client.Do(opt)
	if err != nil {
		return nil, errors.WithMessagef(err, "请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessagef(err, "读取响应失败")
	}
	value, err := cli.handlePrometheusResponse(resp, body)
	if err != nil {
		return nil, err
	}
	return convertMetricPointsForPrometheusModel(value), nil
}

// QueryRange 查询范围数据
func (cli *Client) QueryRange(query string, start, end time.Time, step string) ([]*datasource.MetricSeries, error) {
	uri := fmt.Sprintf("%s/query_range", apiPrefix)
	opt := httpx.NewRequestOption(
		httpx.WithQueryParam("query", query),
		httpx.WithQueryParam("start", formatTime(start)),
		httpx.WithQueryParam("end", formatTime(end)),
		httpx.WithQueryParam("step", step),
		httpx.WithHeaders(cli.customHeaders),
		httpx.WithMethod(http.MethodGet),
		httpx.WithPath(uri),
	)
	resp, err := cli.client.Do(opt)
	if err != nil {
		return nil, errors.WithMessagef(err, "请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessagef(err, "读取响应失败")
	}
	value, err := cli.handlePrometheusResponse(resp, body)
	if err != nil {
		return nil, err
	}
	return convertMetricSeriesForPrometheusModel(value), nil
}

// Write 写入数据
func (cli *Client) Write(timeSeries []prompb.TimeSeries) error {
	req := &prompb.WriteRequest{
		Timeseries: timeSeries,
	}
	tsData, err := proto.Marshal(req)
	if err != nil {
		return errors.WithMessagef(err, "序列化timeSeries失败")
	}
	// 使用snappy压缩
	compressed := snappy.Encode(nil, tsData)
	var headers = make(map[string]string)
	headers["Content-Type"] = "application/x-protobuf"
	headers["Content-Encoding"] = "snappy"
	headers["X-Prometheus-Remote-Write-Version"] = "0.1.0"
	uri := fmt.Sprintf("%s/write", apiPrefix)

	opt := httpx.NewRequestOption(
		httpx.WithBody(compressed),
		httpx.WithHeaders(cli.customHeaders),
		httpx.WithMethod(http.MethodPost),
		httpx.WithPath(uri),
	)
	resp, err := cli.client.Do(opt)
	if err != nil {
		return errors.WithMessagef(err, "请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithMessagef(err, "读取响应失败")
	}
	logs.Debugf("请求返回：%s", string(body))
	statusCode := resp.StatusCode
	if !apiError(statusCode) && statusCode/100 != 2 {
		return errors.Errorf("请求失败，状态码：%d", statusCode)
	}
	return nil
}
