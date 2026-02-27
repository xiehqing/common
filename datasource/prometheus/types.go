package prometheus

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/xiehqing/common/datasource"
	"github.com/xiehqing/common/pkg/util"
	"math"
	"net/http"
	"strings"
	"time"
)

// ParsePromQl 解析promql
func ParsePromQl(promql string) error {
	_, err := parser.ParseExpr(promql)
	return err
}

// ExtractPromQlFromExpr 从promql中提取指标名称
func ExtractPromQlFromExpr(input string) ([]string, error) {
	expr, err := parser.ParseExpr(input)
	if err != nil {
		return nil, errors.WithMessagef(err, "promql格式有误.")
	}
	var lst []string
	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if vs, ok := node.(*parser.VectorSelector); ok {
			lst = append(lst, vs.Name)
		}
		return nil
	})
	lst = util.RemoveDuplicates(lst)
	return lst, nil
}

// ModifyPromQlWithLabels 修改promql的标签
func ModifyPromQlWithLabels(promql string, labelKeyValues map[string]string) (string, error) {
	expr, err := parser.ParseExpr(promql)
	if err != nil {
		return "", errors.WithMessagef(err, "promql格式有误.")
	}

	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if vs, ok := node.(*parser.VectorSelector); ok {
			// 1. 先用 map 记录已有标签
			existing := make(map[string]*labels.Matcher)
			for _, matcher := range vs.LabelMatchers {
				existing[matcher.Name] = matcher
			}
			// 2. 保留原有标签，补充缺失标签
			for k, v := range labelKeyValues {
				if _, ok := existing[k]; !ok {
					matcher, _ := labels.NewMatcher(labels.MatchEqual, k, v)
					vs.LabelMatchers = append(vs.LabelMatchers, matcher)
				}
			}
		}
		return nil
	})

	return expr.String(), nil
}

const (
	statusAPIError = 422
	apiPrefix      = "/api/v1"
)

func apiError(code int) bool {
	return code == statusAPIError || code == http.StatusBadRequest
}

// Config prometheus配置
type Config struct {
	Url                 string            `json:"url"`
	Username            string            `json:"username"`
	Password            string            `json:"password"`
	Headers             map[string]string `json:"headers"`
	Timeout             int               `json:"timeout"`             //请求响应超时时间(毫秒)
	DailTimeout         int               `json:"dailTimeout"`         // 连接建立超时时间(毫秒)
	MaxIdleConnsPerHost int               `json:"maxIdleConnsPerHost"` // 每个主机最大空闲连接数
}

// Validate 校验
func (c *Config) Validate() error {
	if c.Url == "" {
		return errors.Errorf("prometheus url is empty")
	}
	return nil
}

type Range struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
type QueryParam struct {
	Query string     `json:"query"`
	Time  *time.Time `json:"time"`
	Range *Range     `json:"range"`
	Step  string
}

// Validate 校验
func (q *QueryParam) Validate() error {
	if q.Query == "" {
		return errors.Errorf("query is empty")
	}
	if q.Time == nil && q.Range == nil {
		return errors.Errorf("timestamp or range is empty")
	}
	if q.Time != nil {
		if q.Step == "" {
			q.Step = "2m"
		}
	} else {
		if q.Step == "" {
			q.Step = getStepFromDuration(q.Range.End.Sub(q.Range.Start).Seconds())
		}
	}
	return nil
}

type Response struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType string          `json:"errorType"`
	Error     string          `json:"error"`
	Warnings  []string        `json:"warnings,omitempty"`
}

type QueryResult struct {
	ResultType model.ValueType `json:"resultType"`
	Result     interface{}     `json:"result"`
	v          model.Value
}

func (qr *QueryResult) UnmarshalJSON(data []byte) error {
	v := struct {
		Type   model.ValueType `json:"resultType"`
		Result json.RawMessage `json:"result"`
	}{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	switch v.Type {
	case model.ValScalar:
		var sv model.Scalar
		err = json.Unmarshal(v.Result, &sv)
		qr.v = &sv

	case model.ValVector:
		var vv model.Vector
		err = json.Unmarshal(v.Result, &vv)
		qr.v = vv

	case model.ValMatrix:
		var mv model.Matrix
		err = json.Unmarshal(v.Result, &mv)
		qr.v = mv

	default:
		err = fmt.Errorf("unexpected value type %q", v.Type)
	}
	return err
}

// formatTime 格式化时间
func formatTime(t time.Time) string {
	milli := t.UnixMilli()
	toString := util.ToString(milli)
	return fmt.Sprintf("%s.%s", toString[0:10], toString[10:])
}

func roundToMilliseconds(num float64) float64 {
	return math.Round(num*1000) / 1000
}

// getStepFromDuration 根据持续时间获取步长
func getStepFromDuration(duration float64) string {
	result := roundToMilliseconds(duration)
	integerStep := math.Round(duration)
	if duration >= 100 {
		result = integerStep - math.Mod(integerStep, 10)
	} else if duration < 100 && duration >= 10 {
		result = integerStep - math.Mod(integerStep, 5)
	} else if duration < 10 && duration >= 1 {
		result = integerStep
	} else if duration < 1 && duration > 0.01 {
		result = math.Round(duration*40) / 40
	}
	if result <= 0 {
		result = 0.001
	}
	humanized := humanizeSeconds(result)
	// 移除空格
	return strings.ReplaceAll(humanized, " ", "")
}

// humanizeSeconds 将秒数转换为人类可读格式
func humanizeSeconds(seconds float64) string {
	if seconds == 0 {
		return "0s"
	}
	duration := time.Duration(seconds * float64(time.Second))
	// 转换为各个时间单位
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	milliseconds := int(duration.Nanoseconds()/1000000) % 1000
	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if secs > 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}
	if milliseconds > 0 && len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dms", milliseconds))
	}
	if len(parts) == 0 {
		return "1ms"
	}
	return strings.Join(parts, "")
}

// covertMetricToMap 将prometheus的metric转换为map
func covertMetricToMap(metric model.Metric) *datasource.Metric {
	val := &datasource.Metric{}
	for key, value := range metric {
		val.Set(string(key), string(value))
	}
	return val
}

// convertMetricPointsForPrometheusModel 将普米模型数据转换为指标点
func convertMetricPointsForPrometheusModel(value model.Value) []*datasource.MetricPoint {
	if value == nil {
		return nil
	}
	var lst []*datasource.MetricPoint
	switch value.Type() {
	case model.ValVector:
		items, ok := value.(model.Vector)
		if !ok {
			return nil
		}
		for _, item := range items {
			if math.IsNaN(float64(item.Value)) {
				continue
			}
			mp := &datasource.MetricPoint{
				Key:    item.Metric.String(),
				Labels: covertMetricToMap(item.Metric),
			}
			mp.Timestamp = item.Timestamp.Unix()
			mp.Value = float64(item.Value)
			lst = append(lst, mp)
		}
	case model.ValMatrix:
		items, ok := value.(model.Matrix)
		if !ok {
			return nil
		}
		for _, item := range items {
			if len(item.Values) == 0 {
				return nil
			}
			last := item.Values[len(item.Values)-1]
			if math.IsNaN(float64(last.Value)) {
				continue
			}
			mp := &datasource.MetricPoint{
				Key:    item.Metric.String(),
				Labels: covertMetricToMap(item.Metric),
			}
			mp.Timestamp = last.Timestamp.Unix()
			mp.Value = float64(last.Value)
			lst = append(lst, mp)
		}
	case model.ValScalar:
		item, ok := value.(*model.Scalar)
		if !ok {
			return nil
		}
		if math.IsNaN(float64(item.Value)) {
			return nil
		}
		mp := &datasource.MetricPoint{
			Key:    "{}",
			Labels: &datasource.Metric{},
		}
		mp.Timestamp = item.Timestamp.Unix()
		mp.Value = float64(item.Value)
		lst = append(lst, mp)
	default:
		return lst
	}
	return lst
}

// convertMetricSeriesForPrometheusModel 将普米模型数据转换为指标序列
func convertMetricSeriesForPrometheusModel(value model.Value) []*datasource.MetricSeries {
	if value == nil {
		return nil
	}
	var lst []*datasource.MetricSeries

	switch value.Type() {
	case model.ValVector:
		items, ok := value.(model.Vector)
		if !ok {
			return lst
		}
		for _, item := range items {
			if math.IsNaN(float64(item.Value)) {
				continue
			}
			timeVal := datasource.DataPoint{
				Timestamp: item.Timestamp.Unix(),
				Value:     float64(item.Value),
			}
			ms := &datasource.MetricSeries{
				Key:    item.Metric.String(),
				Labels: covertMetricToMap(item.Metric),
			}
			ms.DataPoints = append(ms.DataPoints, &timeVal)
			lst = append(lst, ms)
		}
	case model.ValMatrix:
		items, ok := value.(model.Matrix)
		if !ok {
			return lst
		}
		for _, item := range items {
			if len(item.Values) == 0 {
				return lst
			}
			last := item.Values[len(item.Values)-1]

			var timeVals []*datasource.DataPoint
			for _, iv := range item.Values {
				timeVals = append(timeVals, &datasource.DataPoint{
					Timestamp: iv.Timestamp.Unix(),
					Value:     float64(iv.Value),
				})
			}

			if math.IsNaN(float64(last.Value)) {
				continue
			}
			ms := &datasource.MetricSeries{
				Key:    item.Metric.String(),
				Labels: covertMetricToMap(item.Metric),
			}
			ms.DataPoints = timeVals
			lst = append(lst, ms)
		}
	case model.ValScalar:
		item, ok := value.(*model.Scalar)
		if !ok {
			return lst
		}
		if math.IsNaN(float64(item.Value)) {
			return lst
		}
		var timeVals = []*datasource.DataPoint{
			{
				Timestamp: item.Timestamp.Unix(),
				Value:     float64(item.Value),
			},
		}
		ms := &datasource.MetricSeries{
			Key:    "{}",
			Labels: &datasource.Metric{},
		}
		ms.DataPoints = timeVals

		lst = append(lst, ms)
	default:
		return lst
	}
	return lst

}
