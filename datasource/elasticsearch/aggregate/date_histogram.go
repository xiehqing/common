package aggregate

import (
	"github.com/olivere/elastic"
)

type DateHistogram struct{}

func (h *DateHistogram) Name() string {
	return "date_histogram"
}

func (h *DateHistogram) Aggregation(opt *Option) *Obj {
	aggregation := elastic.NewDateHistogramAggregation().
		Field(opt.Field).
		Interval(opt.DateHistogramInterval)
	if opt.TimeZone != "" {
		aggregation = aggregation.TimeZone(opt.TimeZone)
	} else {
		aggregation = aggregation.TimeZone("Asia/Shanghai")
	}
	if opt.CountDesc {
		aggregation = aggregation.OrderByCountDesc()
	} else {
		aggregation = aggregation.OrderByCountAsc()
	}
	if opt.KeyDesc {
		aggregation = aggregation.OrderByKeyDesc()
	} else {
		aggregation = aggregation.OrderByKeyAsc()
	}
	return &Obj{
		Name:        opt.Name,
		Aggregation: aggregation,
	}
}
