package request

import (
	"github.com/olivere/elastic"
)

type BoolRequest struct {
	Occur              Occur
	requests           []Request
	MinimumShouldMatch string
}

func (b *BoolRequest) Name() string {
	return "bool"
}

func (b *BoolRequest) Query() elastic.Query {
	boolQuery := elastic.NewBoolQuery()
	if len(b.requests) > 0 {
		var queries []elastic.Query
		for _, req := range b.requests {
			queries = append(queries, req.Query())
		}
		switch b.Occur {
		case Must:
			boolQuery = boolQuery.Must(queries...)
		case Should:
			boolQuery = boolQuery.Should(queries...)
			if b.MinimumShouldMatch != "" {
				boolQuery = boolQuery.MinimumShouldMatch(b.MinimumShouldMatch)
			}
		case MustNot:
			boolQuery = boolQuery.MustNot(queries...)
		case Filter:
			boolQuery = boolQuery.Filter(queries...)
		}
	}
	return boolQuery
}

type BoolRequestBuilder struct {
	boolRequest BoolRequest
}

func BoolBuilder() *BoolRequestBuilder {
	return &BoolRequestBuilder{
		boolRequest: BoolRequest{
			requests: make([]Request, 0),
		},
	}
}

func (b *BoolRequestBuilder) Build() BoolRequest {
	return b.boolRequest
}

func (b *BoolRequestBuilder) SetOccur(occur Occur) *BoolRequestBuilder {
	b.boolRequest.Occur = occur
	return b
}

func (b *BoolRequestBuilder) AddRequest(req Request) *BoolRequestBuilder {
	b.boolRequest.requests = append(b.boolRequest.requests, req)
	return b
}

func (b *BoolRequestBuilder) SetMinimumShouldMatch(minimumShouldMatch string) *BoolRequestBuilder {
	b.boolRequest.MinimumShouldMatch = minimumShouldMatch
	return b
}

func (b *BoolRequestBuilder) SetRequests(requests []Request) *BoolRequestBuilder {
	b.boolRequest.requests = requests
	return b
}
