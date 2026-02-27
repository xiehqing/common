package aggregate

import (
	"github.com/olivere/elastic"
)

type Min struct {
}

func (a *Min) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewMinAggregation().Field(agg.Field),
	}
}

func (a *Min) Name() string {
	return "min"
}
