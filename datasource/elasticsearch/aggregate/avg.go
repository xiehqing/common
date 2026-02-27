package aggregate

import (
	"github.com/olivere/elastic"
)

type Avg struct {
}

func (a *Avg) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewAvgAggregation().Field(agg.Field),
	}
}

func (a *Avg) Name() string {
	return "avg"
}
