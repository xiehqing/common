package aggregate

import (
	"github.com/olivere/elastic"
)

type Max struct {
}

func (a *Max) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewMaxAggregation().Field(agg.Field),
	}
}

func (a *Max) Name() string {
	return "max"
}
