package request

import "github.com/olivere/elastic"

type TermsRequest struct {
	Field string
	Value []interface{}
}

func (tr *TermsRequest) Name() string {
	return "terms"
}

func (tr *TermsRequest) Query() elastic.Query {
	return elastic.NewTermsQuery(tr.Field, tr.Value...)
}

type TermsRequestBuilder struct {
	terms TermsRequest
}

func TermsBuilder() *TermsRequestBuilder {
	return &TermsRequestBuilder{
		terms: TermsRequest{
			Value: make([]interface{}, 0),
		},
	}
}

func (t *TermsRequestBuilder) SetField(field string) *TermsRequestBuilder {
	t.terms.Field = field
	return t
}

func (t *TermsRequestBuilder) SetValue(value []interface{}) *TermsRequestBuilder {
	t.terms.Value = value
	return t
}

func (t *TermsRequestBuilder) AddValue(value interface{}) *TermsRequestBuilder {
	t.terms.Value = append(t.terms.Value, value)
	return t
}

func (t *TermsRequestBuilder) Build() TermsRequest {
	return t.terms
}
