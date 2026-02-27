package aggregate

import (
	"github.com/olivere/elastic"
)

type TopHits struct{}

func (a *TopHits) Name() string {
	return "top_hits"
}

func (a *TopHits) Aggregation(opt *Option) *Obj {
	return &Obj{
		Name: opt.Name,
		Aggregation: elastic.NewTopHitsAggregation().
			Sort(opt.TopHitsSortField, !opt.TopHitsSortDesc).
			Size(opt.TopSize),
	}
}
