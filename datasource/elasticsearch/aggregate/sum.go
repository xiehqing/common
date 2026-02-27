package aggregate

import (
	"github.com/olivere/elastic"
)

type Sum struct {
}

func (a *Sum) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewSumAggregation().Field(agg.Field),
	}
}

func (a *Sum) Name() string {
	return "sum"
}
