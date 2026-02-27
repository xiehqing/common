package aggregate

import (
	"github.com/olivere/elastic"
)

type Cardinality struct{}

func (a *Cardinality) Name() string {
	return "cardinality"
}

func (a *Cardinality) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewCardinalityAggregation().Field(agg.Field),
	}
}
