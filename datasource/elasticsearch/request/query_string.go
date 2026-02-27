package request

import "github.com/olivere/elastic"

type QueryStringRequest struct {
	QueryString        string
	Operator           string
	MinimumShouldMatch string
	fields             map[string]float64
}

func (q *QueryStringRequest) Name() string {
	return "query_string"
}

func (q *QueryStringRequest) Query() elastic.Query {
	query := elastic.NewQueryStringQuery(q.QueryString)
	if q.Operator != "" {
		query = query.DefaultOperator(q.Operator)
	}
	if q.MinimumShouldMatch != "" {
		query = query.MinimumShouldMatch(q.MinimumShouldMatch)
	}
	if q.fields != nil {
		for field, boost := range q.fields {
			query = query.FieldWithBoost(field, boost)
		}
	}
	return query
}

type QueryStringRequestBuilder struct {
	queryStringRequest QueryStringRequest
}

func QueryStringBuilder() *QueryStringRequestBuilder {
	return &QueryStringRequestBuilder{
		queryStringRequest: QueryStringRequest{},
	}
}

func (q *QueryStringRequestBuilder) Build() QueryStringRequest {
	return q.queryStringRequest
}

func (q *QueryStringRequestBuilder) SetQueryString(queryString string) *QueryStringRequestBuilder {
	q.queryStringRequest.QueryString = queryString
	return q
}

func (q *QueryStringRequestBuilder) SetOperator(operator string) *QueryStringRequestBuilder {
	q.queryStringRequest.Operator = operator
	return q
}

func (q *QueryStringRequestBuilder) SetMinimumShouldMatch(minimumShouldMatch string) *QueryStringRequestBuilder {
	q.queryStringRequest.MinimumShouldMatch = minimumShouldMatch
	return q
}

func (q *QueryStringRequestBuilder) SetFields(fields map[string]float64) *QueryStringRequestBuilder {
	q.queryStringRequest.fields = fields
	return q
}

func (q *QueryStringRequestBuilder) AddField(field string, boost float64) *QueryStringRequestBuilder {
	if q.queryStringRequest.fields == nil {
		q.queryStringRequest.fields = make(map[string]float64)
	}
	q.queryStringRequest.fields[field] = boost
	return q
}
