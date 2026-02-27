package elasticsearch

import (
	"github.com/olivere/elastic"
	"github.com/xiehqing/common/datasource/elasticsearch/aggregate"
	"github.com/xiehqing/common/datasource/elasticsearch/request"
	"github.com/xiehqing/common/pkg/logs"
	"reflect"
)

type Config struct {
	Uris           []string `json:"uris" yaml:"uris" mapstructure:"uris"`
	Username       string   `json:"username" yaml:"username" mapstructure:"username"`
	Password       string   `json:"password" yaml:"password" mapstructure:"password"`
	Timeout        int64    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	SkipTlsVerify  bool     `json:"skipTlsVerify" yaml:"skip-tls-verify" mapstructure:"skip-tls-verify"`
	Version        string   `json:"version" yaml:"version" mapstructure:"version"`
	EnableTraceLog bool     `json:"enableTraceLog" yaml:"enable-trace-log" mapstructure:"enable-trace-log"`
}

var aggregatorCache = map[string]aggregate.Aggregate{}

func RegisterAggregator() {
	aggregatorCache["histogram"] = &aggregate.Histogram{}
	aggregatorCache["date_histogram"] = &aggregate.DateHistogram{}
	aggregatorCache["percentiles"] = &aggregate.Percentiles{}
	aggregatorCache["top_hits"] = &aggregate.TopHits{}
	aggregatorCache["max"] = &aggregate.Max{}
	aggregatorCache["min"] = &aggregate.Min{}
	aggregatorCache["sum"] = &aggregate.Sum{}
	aggregatorCache["avg"] = &aggregate.Avg{}
	aggregatorCache["cardinality"] = &aggregate.Cardinality{}
	aggregatorCache["count"] = &aggregate.Count{}
	aggregatorCache["terms"] = &aggregate.Terms{}
}

type Sort struct {
	Field string
	Desc  bool
}

type RequestOption struct {
	Index              string
	Type               string
	Desc               bool
	Sort               string
	UnmappedType       string
	PageSize           int
	PageNum            int
	Must               []request.Request
	Should             []request.Request
	MustNot            []request.Request
	Filter             []request.Request
	MultiSort          bool
	Sorts              []elastic.FieldSort
	MinimumShouldMatch string
	IncludeFields      []string
	ExcludeFields      []string
	AggReqs            []aggregate.Option
	ScrollKeepAlive    string
}

func (opt *RequestOption) PreHandle() {
	if opt.UnmappedType == "" {
		opt.UnmappedType = "string"
	}
}

func BuildQuery(opt *RequestOption) elastic.Query {
	boolQuery := elastic.NewBoolQuery()
	if len(opt.Must) > 0 {
		var queries []elastic.Query
		for _, req := range opt.Must {
			queries = append(queries, req.Query())
		}
		boolQuery = boolQuery.Must(queries...)
	}
	if len(opt.Should) > 0 {
		var queries []elastic.Query
		for _, req := range opt.Should {
			queries = append(queries, req.Query())
		}
		boolQuery = boolQuery.Should(queries...)
		if opt.MinimumShouldMatch != "" {
			boolQuery = boolQuery.MinimumShouldMatch(opt.MinimumShouldMatch)
		}
	}
	if len(opt.MustNot) > 0 {
		var queries []elastic.Query
		for _, req := range opt.MustNot {
			queries = append(queries, req.Query())
		}
		boolQuery = boolQuery.MustNot(queries...)
	}
	if len(opt.Filter) > 0 {
		var queries []elastic.Query
		for _, req := range opt.Filter {
			queries = append(queries, req.Query())
		}
		boolQuery = boolQuery.Filter(queries...)
	}
	return boolQuery
}

func BuildSort(opt *RequestOption) []elastic.Sorter {
	opt.PreHandle()
	var sorts []elastic.Sorter
	if opt.MultiSort {
		if len(opt.Sorts) == 0 {
			for _, sort := range opt.Sorts {
				sorts = append(sorts, sort.Sorter)
			}
		}
	} else {
		if opt.Desc {
			sorts = append(sorts, elastic.NewFieldSort(opt.Sort).UnmappedType(opt.UnmappedType).Desc())
		} else {
			sorts = append(sorts, elastic.NewFieldSort(opt.Sort).UnmappedType(opt.UnmappedType).Asc())
		}
	}
	return sorts
}

func BuildFetchSourceContext(opt *RequestOption) *elastic.FetchSourceContext {
	var fetchSource = false
	if len(opt.IncludeFields) == 0 && len(opt.ExcludeFields) == 0 {
		fetchSource = true
	}
	return elastic.NewFetchSourceContext(fetchSource).Include(opt.IncludeFields...).Exclude(opt.ExcludeFields...)
}

// BuildAggregate 构建聚合查询
func BuildAggregate(service *elastic.SearchService, opt aggregate.Option) {
	aggregation := buildAggregation(opt, opt.SubAggReqs)
	if aggregation != nil {
		service = service.Aggregation(opt.Name, aggregation)
	}
}

// buildAggregation 递归构建聚合查询
func buildAggregation(opt aggregate.Option, subOpts []aggregate.Option) elastic.Aggregation {
	typ := opt.Type
	if aggregator, ok := aggregatorCache[typ]; ok {
		aggregation := aggregator.Aggregation(&opt)
		// 使用反射获取聚合类型的SubAggregation方法，如果不存在此方法，则不支持添加子聚合
		aggValue := reflect.ValueOf(aggregation.Aggregation)
		var hasSubAggMethod = false
		var subAggMethod reflect.Method
		for i := 0; i < aggValue.NumMethod(); i++ {
			method := aggValue.Type().Method(i)
			if method.Name == "SubAggregation" {
				hasSubAggMethod = true
				subAggMethod = method
			}
		}
		if hasSubAggMethod {
			for _, agg := range subOpts {
				// 使用反射调用SubAggregation方法追加子聚合
				subAggMethod.Func.Call([]reflect.Value{reflect.ValueOf(aggregation.Aggregation), reflect.ValueOf(agg.Name), reflect.ValueOf(buildAggregation(agg, agg.SubAggReqs))})
			}
		} else {
			logs.Errorf("%s聚合类型下不支持添加子聚合.", opt.Type)
		}
		return aggregation.Aggregation
	} else {
		logs.Errorf("暂不支持的聚合类型:%s", opt.Type)
	}
	return nil
}
