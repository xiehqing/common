package aggregate

import (
	"github.com/olivere/elastic"
)

type Terms struct{}

func (a *Terms) Name() string {
	return "terms"
}

func (a *Terms) Aggregation(opt *Option) *Obj {

	aggregation := elastic.NewTermsAggregation()
	if opt.Script != nil {
		aggregation = aggregation.Script(opt.Script)
	} else {
		aggregation = aggregation.Field(opt.Field)
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

	if opt.TopSize > 0 {
		aggregation = aggregation.Size(opt.TopSize)
	}

	return &Obj{
		Name:        opt.Name,
		Aggregation: aggregation,
	}
}
