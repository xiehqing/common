package aggregate

import (
	"github.com/olivere/elastic"
)

type Count struct {
}

func (a *Count) Aggregation(agg *Option) *Obj {
	return &Obj{
		Name:        agg.Name,
		Aggregation: elastic.NewValueCountAggregation().Field(agg.Field),
	}
}

func (a *Count) Name() string {
	return "count"
}
