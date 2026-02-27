package aggregate

import (
	"github.com/olivere/elastic"
)

type Histogram struct{}

func (h *Histogram) Name() string {
	return "histogram"
}

func (h *Histogram) Aggregation(opt *Option) *Obj {
	aggregation := elastic.NewHistogramAggregation().Field(opt.Field).Interval(opt.HistogramInterval)
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
