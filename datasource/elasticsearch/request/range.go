package request

import "github.com/olivere/elastic"

type RangeRequest struct {
	RangeName    string
	From         interface{}
	To           interface{}
	IncludeLower bool
	IncludeUpper bool
}

func (rr *RangeRequest) Name() string {
	return "range"
}

func (rr *RangeRequest) Query() elastic.Query {
	rangeQuery := elastic.NewRangeQuery(rr.RangeName)
	if rr.From != nil {
		rangeQuery = rangeQuery.From(rr.From)
	}
	if rr.To != nil {
		rangeQuery = rangeQuery.To(rr.To)
	}
	rangeQuery = rangeQuery.IncludeLower(rr.IncludeLower).IncludeUpper(rr.IncludeUpper)
	return rangeQuery
}

type RangeRequestBuilder struct {
	rangeRequest RangeRequest
}

func RangeBuilder() *RangeRequestBuilder {
	return &RangeRequestBuilder{
		rangeRequest: RangeRequest{
			IncludeLower: true,
			IncludeUpper: true,
		},
	}
}

func (r *RangeRequestBuilder) Build() RangeRequest {
	return r.rangeRequest
}

func (r *RangeRequestBuilder) SetRangeName(rangeName string) *RangeRequestBuilder {
	r.rangeRequest.RangeName = rangeName
	return r
}

func (r *RangeRequestBuilder) SetFrom(from interface{}) *RangeRequestBuilder {
	r.rangeRequest.From = from
	return r
}

func (r *RangeRequestBuilder) SetTo(to interface{}) *RangeRequestBuilder {
	r.rangeRequest.To = to
	return r
}

func (r *RangeRequestBuilder) SetIncludeLower(includeLower bool) *RangeRequestBuilder {
	r.rangeRequest.IncludeLower = includeLower
	return r
}

func (r *RangeRequestBuilder) SetIncludeUpper(includeUpper bool) *RangeRequestBuilder {
	r.rangeRequest.IncludeUpper = includeUpper
	return r
}
