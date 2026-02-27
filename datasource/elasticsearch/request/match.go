package request

import "github.com/olivere/elastic"

type MatchRequest struct {
	Field              string
	Text               interface{}
	Operator           string
	MinimumShouldMatch string
}

func (m *MatchRequest) Name() string {
	return "match"
}
func (m *MatchRequest) Query() elastic.Query {
	query := elastic.NewMatchQuery(m.Field, m.Text)
	if m.Operator != "" {
		query = query.Operator(m.Operator)
	}
	if m.MinimumShouldMatch != "" {
		query = query.MinimumShouldMatch(m.MinimumShouldMatch)
	}
	return query
}

type MatchRequestBuilder struct {
	match MatchRequest
}

func MatchBuilder() *MatchRequestBuilder {
	return &MatchRequestBuilder{}
}

func (m *MatchRequestBuilder) Build() MatchRequest {
	return m.match
}

func (m *MatchRequestBuilder) SetField(field string) *MatchRequestBuilder {
	m.match.Field = field
	return m
}

func (m *MatchRequestBuilder) SetText(text interface{}) *MatchRequestBuilder {
	m.match.Text = text
	return m
}

func (m *MatchRequestBuilder) SetOperator(operator string) *MatchRequestBuilder {
	m.match.Operator = operator
	return m
}

func (m *MatchRequestBuilder) SetMinimumShouldMatch(minimumShouldMatch string) *MatchRequestBuilder {
	m.match.MinimumShouldMatch = minimumShouldMatch
	return m
}
