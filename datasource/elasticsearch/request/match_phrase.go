package request

import "github.com/olivere/elastic"

type MatchPhraseRequest struct {
	Field string
	Text  interface{}
}

func (m *MatchPhraseRequest) Name() string {
	return "match_phrase"
}

func (m *MatchPhraseRequest) Query() elastic.Query {
	return elastic.NewMatchPhraseQuery(m.Field, m.Text)
}

type MatchPhraseRequestBuilder struct {
	matchPhrase MatchPhraseRequest
}

func MatchPhraseBuilder() *MatchPhraseRequestBuilder {
	return &MatchPhraseRequestBuilder{}
}

func (m *MatchPhraseRequestBuilder) Build() MatchPhraseRequest {
	return m.matchPhrase
}

func (m *MatchPhraseRequestBuilder) SetField(field string) *MatchPhraseRequestBuilder {
	m.matchPhrase.Field = field
	return m
}

func (m *MatchPhraseRequestBuilder) SetText(text interface{}) *MatchPhraseRequestBuilder {
	m.matchPhrase.Text = text
	return m
}
