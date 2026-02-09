package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hatcher/common/pkg/logs"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client 通用HTTP客户端工具
type Client struct {
	Client  *http.Client
	BaseUrl string
}

// NewClient 创建一个新的HTTPClient实例
func NewClient(baseUrl string, timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: timeout,
		},
		BaseUrl: baseUrl,
	}
}

// NewTransportClient 创建一个新的HTTPClient实例
func NewTransportClient(baseUrl string, transport http.RoundTripper, timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		BaseUrl: baseUrl,
	}
}

// NewDefaultClient 创建一个新的HTTPClient实例，默认超时时间为10秒
func NewDefaultClient(baseUrl string) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseUrl: baseUrl,
	}
}

// buildRequest 构建HTTP请求
func (c *Client) buildRequest(options *RequestOption) (*http.Request, error) {
	var body io.Reader
	if options.Body != nil {
		// 判断options.Body是否是[]byte
		if _, ok := options.Body.([]byte); ok {
			body = bytes.NewBuffer(options.Body.([]byte))
		} else {
			jsonData, err := json.Marshal(options.Body)
			if err != nil {
				return nil, fmt.Errorf("解析http请求结果失败: %v", err)
			}
			body = bytes.NewBuffer(jsonData)
		}
	}
	// 处理查询参数
	reqURL := c.BaseUrl + options.Path
	if len(options.Query) > 0 {
		params := url.Values{}
		for key, value := range options.Query {
			params.Add(key, value)
		}
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}
	req, err := http.NewRequest(options.Method.String(), reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	// 设置请求头
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

// Do 发送HTTP请求并返回响应
func (c *Client) Do(options *RequestOption) (*http.Response, error) {
	requestTime := time.Now()
	logs.Infof("请求方法: %v", options)
	request, err := c.buildRequest(options)
	logs.Infof("发送HTTP请求: %v", request)
	logs.Infof("打印日志: %v", options.PrintLog)
	if options.PrintLog {
		r := &RequestLog{
			Timestamp: requestTime.Format("2006-01-02 15:04:05.000"),
			Method:    options.Method.String(),
			URL:       request.URL.String(),
			Headers:   options.Headers,
			Body:      options.Body,
			RequestID: options.RequestID,
		}
		// 判断options.Body是否是[]byte
		if _, ok := options.Body.([]byte); ok {
			r.Body = string(options.Body.([]byte))
		} else {
			jsonData, err := json.Marshal(options.Body)
			if err != nil {
				logs.Errorf("解析http请求结果失败: %v", err)
			} else {
				r.Body = string(jsonData)
			}
		}
		LogRequestJSON(r, options.Sensitive)
	}
	if err != nil {
		return nil, err
	}
	response, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 读取响应体内容到缓冲区
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// 创建可以重复读取的响应
	bufferedResp := &BufferedResponse{
		Response: &http.Response{
			Status:           response.Status,
			StatusCode:       response.StatusCode,
			Proto:            response.Proto,
			ProtoMajor:       response.ProtoMajor,
			ProtoMinor:       response.ProtoMinor,
			Header:           response.Header,
			ContentLength:    int64(len(bodyBytes)),
			TransferEncoding: response.TransferEncoding,
			Close:            response.Close,
			Uncompressed:     response.Uncompressed,
			Trailer:          response.Trailer,
			Request:          response.Request,
			TLS:              response.TLS,
		},
		bodyBuffer: bodyBytes,
	}

	// 设置可以重复读取的Body
	bufferedResp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	responseTime := time.Now()
	if options.PrintLog {
		r := &ResponseLog{
			Timestamp:  responseTime.Format("2006-01-02 15:04:05.000"),
			StatusCode: response.StatusCode,
			RequestID:  options.RequestID,
			DurationMs: responseTime.Sub(requestTime).Milliseconds(),
			Body:       string(bodyBytes),
		}
		if err != nil {
			r.Error = err.Error()
		}
		LogResponseJSON(r)
	}
	return bufferedResp.Response, nil
}

// DoWithPtr 发送HTTP请求并返回响应
func (c *Client) DoWithPtr(options *RequestOption, resp interface{}) error {
	response, err := c.Do(options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// 读取响应体内容到缓冲区
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, resp)
}
