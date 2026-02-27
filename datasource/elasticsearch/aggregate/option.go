package aggregate

import "github.com/olivere/elastic"

type Aggregate interface {
	Name() string
	Aggregation(agg *Option) *Obj
}

type Option struct {
	Name                  string
	Field                 string
	Type                  string
	CountDesc             bool
	KeyDesc               bool
	TopSize               int
	TopHitsSortField      string
	TopHitsSortDesc       bool
	HistogramInterval     float64
	DateHistogramInterval string
	TimeZone              string
	Script                *elastic.Script
	Percentiles           []float64
	SubAggReqs            []Option
}

type Obj struct {
	Name        string
	Aggregation elastic.Aggregation
}
