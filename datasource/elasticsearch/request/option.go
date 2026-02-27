package request

import "github.com/olivere/elastic"

type Request interface {
	Name() string
	Query() elastic.Query
}

type Occur string

const (
	Must    Occur = "must"
	Should  Occur = "should"
	MustNot Occur = "must_not"
	Filter  Occur = "filter"
)
