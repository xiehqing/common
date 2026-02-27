package datasource

import (
	"fmt"
	"github.com/prometheus/prometheus/prompb"
	"github.com/toolkits/pkg/str"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MetricNameLabel = "__name__"
)

type LabelName string

func (ln LabelName) IsValid() bool {
	if len(ln) == 0 {
		return false
	}
	for i, b := range ln {
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || (b >= '0' && b <= '9' && i > 0)) {
			return false
		}
	}
	return true
}

type LabelValue string

func (lv LabelValue) IsValid() bool {
	return utf8.ValidString(string(lv))
}

type LabelSet map[LabelName]LabelValue

func (ls LabelSet) String() string {
	rsLst := make([]string, 0, len(ls))
	for l, v := range ls {
		rsLst = append(rsLst, fmt.Sprintf("%s=%q", l, v))
	}
	sort.Strings(rsLst)
	return fmt.Sprintf("{%s}", strings.Join(rsLst, ", "))
}

type Metric LabelSet

func (m Metric) Set(key, value string) {
	m[LabelName(key)] = LabelValue(value)
}

func (m Metric) Get(key string) LabelValue {
	return m[LabelName(key)]
}

func (m Metric) String() string {
	metricName, hasName := m[MetricNameLabel]
	numLabels := len(m) - 1
	if !hasName {
		numLabels = len(m)
	}
	labelStrings := make([]string, 0, numLabels)
	for label, value := range m {
		if label != MetricNameLabel {
			labelStrings = append(labelStrings, fmt.Sprintf("%s=%q", label, value))
		}
	}

	switch numLabels {
	case 0:
		if hasName {
			return string(metricName)
		}
		return "{}"
	default:
		sort.Strings(labelStrings)
		return fmt.Sprintf("%s{%s}", metricName, strings.Join(labelStrings, ", "))
	}
}

func (m Metric) Hash() string {
	return str.MD5(m.String())
}

func (m Metric) ToLabelMap() map[string]string {
	labelMap := make(map[string]string)
	for k, v := range m {
		labelMap[string(k)] = string(v)
	}
	return labelMap
}

func (m Metric) ToPromLabel() []prompb.Label {
	var promLabels []prompb.Label
	for k, v := range m {
		promLabels = append(promLabels, prompb.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	return promLabels
}

type MetricPoint struct {
	DataPoint
	Key    string  `json:"key"`
	Labels *Metric `json:"labels"`
}

type MetricSeries struct {
	Key        string       `json:"key"`
	Labels     *Metric      `json:"labels"`
	DataPoints []*DataPoint `json:"dataPoints"`
}

type DataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type MetricLine struct {
	Key    string    `json:"key"`
	Labels *Metric   `json:"labels"`
	XAxis  []string  `json:"xAxis"`
	YAxis  []float64 `json:"yAxis"`
}

// ConvertSeriesToLine 转换指标序列为折线图
func ConvertSeriesToLine(series []*MetricSeries) []*MetricLine {
	var lines = make([]*MetricLine, 0)
	for _, serie := range series {
		ml := &MetricLine{
			Key:    serie.Key,
			Labels: serie.Labels,
		}
		var xAxis []string
		var yAxis []float64
		for _, dp := range serie.DataPoints {
			xAxis = append(xAxis, time.Unix(dp.Timestamp, 10).Format("2006-01-02 15:04:05"))
			yAxis = append(yAxis, dp.Value)
		}
		ml.XAxis = xAxis
		ml.YAxis = yAxis
		lines = append(lines, ml)
	}
	return lines
}
