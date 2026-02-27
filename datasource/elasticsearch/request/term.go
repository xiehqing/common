package request

import "github.com/olivere/elastic"

type TermRequest struct {
	Field string
	Value interface{}
}

func (tr *TermRequest) Name() string {
	return "term"
}
func (tr *TermRequest) Query() elastic.Query {
	return elastic.NewTermQuery(tr.Field, tr.Value)
}

type TermRequestBuilder struct {
	term TermRequest
}

func TermBuilder() *TermRequestBuilder {
	return &TermRequestBuilder{
		term: TermRequest{},
	}
}

func (t *TermRequestBuilder) Build() TermRequest {
	return t.term
}

func (t *TermRequestBuilder) SetField(field string) *TermRequestBuilder {
	t.term.Field = field
	return t
}

func (t *TermRequestBuilder) SetValue(value interface{}) *TermRequestBuilder {
	t.term.Value = value
	return t
}
