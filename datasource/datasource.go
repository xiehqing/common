package datasource

import "context"

type DataSource interface {
	Init() error
	Equal(ds DataSource) bool
	QueryData(ctx context.Context, query interface{}) ([]*MetricPoint, error)
}
