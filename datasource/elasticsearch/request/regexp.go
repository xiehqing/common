package request

import "github.com/olivere/elastic"

type RegexpRequest struct {
	Field   string
	Pattern string
}

func (r *RegexpRequest) Name() string {
	return "regexp"
}

func (r *RegexpRequest) Query() elastic.Query {
	return elastic.NewRegexpQuery(r.Field, r.Pattern)
}

type RegexpRequestBuilder struct {
	regexpRequest RegexpRequest
}

func RegexpBuilder() *RegexpRequestBuilder {
	return &RegexpRequestBuilder{
		regexpRequest: RegexpRequest{},
	}
}

func (r *RegexpRequestBuilder) Build() RegexpRequest {
	return r.regexpRequest
}

func (r *RegexpRequestBuilder) SetField(field string) *RegexpRequestBuilder {
	r.regexpRequest.Field = field
	return r
}

func (r *RegexpRequestBuilder) SetPattern(pattern string) *RegexpRequestBuilder {
	r.regexpRequest.Pattern = pattern
	return r
}
