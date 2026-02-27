package aggregate

import (
	"github.com/olivere/elastic"
)

type Percentiles struct{}

func (p *Percentiles) Name() string {
	return "percentiles"
}

func (p *Percentiles) Aggregation(opt *Option) *Obj {
	aggregation := elastic.NewPercentilesAggregation().Field(opt.Field)
	if len(opt.Percentiles) > 0 {
		aggregation = aggregation.Percentiles(opt.Percentiles...)
	} else {
		aggregation = aggregation.Percentiles(50.0, 75.0, 90.0, 95.0, 99.0, 100.0)
	}
	return &Obj{
		Name:        opt.Name,
		Aggregation: aggregation,
	}
}
