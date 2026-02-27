package request

import "github.com/olivere/elastic"

type WildcardRequest struct {
	Field string
	Value string
}

func (w *WildcardRequest) Name() string {
	return "wildcard"
}

func (w *WildcardRequest) Query() elastic.Query {
	return elastic.NewWildcardQuery(w.Field, w.Value)
}

type WildcardRequestBuilder struct {
	wildcardRequest WildcardRequest
}

func WildcardBuilder() *WildcardRequestBuilder {
	return &WildcardRequestBuilder{
		wildcardRequest: WildcardRequest{},
	}
}

func (w *WildcardRequestBuilder) Build() WildcardRequest {
	return w.wildcardRequest
}

func (w *WildcardRequestBuilder) SetField(field string) *WildcardRequestBuilder {
	w.wildcardRequest.Field = field
	return w
}

func (w *WildcardRequestBuilder) SetValue(value string) *WildcardRequestBuilder {
	w.wildcardRequest.Value = value
	return w
}
