package elasticsearch

import (
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"github.com/xiehqing/common/pkg/logs"
	"github.com/xiehqing/common/pkg/tlsx"
	"golang.org/x/net/context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Uris     []string
	esClient *elastic.Client
	Version  string
}

type TraceLog struct{}

func (t *TraceLog) Printf(format string, args ...interface{}) {
	logs.Infof(format, args...)
}

// NewElasticClient 创建ElasticSearch客户端
func NewElasticClient(option Config) (*Client, error) {
	if len(option.Uris) == 0 {
		return nil, errors.Errorf("未配置ElasticSearch地址")
	}
	ec := &Client{
		Uris: option.Uris,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: time.Duration(option.Timeout) * time.Millisecond,
		}).DialContext,
		ResponseHeaderTimeout: time.Duration(option.Timeout) * time.Millisecond,
	}

	url := option.Uris[0]
	if strings.Contains(url, "https") {
		tlsConfig := tlsx.ClientConfig{
			InsecureSkipVerify: option.SkipTlsVerify,
			UseTLS:             true,
		}
		cfg, err := tlsConfig.TLSConfig()
		if err != nil {
			return nil, errors.Errorf("初始化ElasticSearch TLS配置失败: %v", err)
		}
		transport.TLSClientConfig = cfg
	}

	options := []elastic.ClientOptionFunc{
		elastic.SetURL(option.Uris...),
	}
	if option.Username != "" {
		options = append(options, elastic.SetBasicAuth(option.Username, option.Password))
	}
	options = append(options, elastic.SetHttpClient(&http.Client{Transport: transport}))
	options = append(options, elastic.SetSniff(false))
	options = append(options, elastic.SetHealthcheck(false))
	if option.EnableTraceLog {
		options = append(options, elastic.SetTraceLog(&TraceLog{}))
	}
	esClient, err := elastic.NewClient(options...)
	if err != nil {
		return nil, errors.Errorf("初始化ElasticSearch客户端失败: %v", err)
	}
	ec.esClient = esClient
	// 初始化聚合工具
	RegisterAggregator()
	return ec, nil
}

// Query 查询
func (ec *Client) Query(ctx context.Context, indices []string, types []string, from, size int, query elastic.Query) (*elastic.SearchResult, error) {
	result, err := ec.esClient.Search().
		Index(indices...).
		Type(types...).
		From(from).
		RestTotalHitsAsInt(true).
		TrackTotalHits(true).
		Size(size).
		Query(query).
		Do(ctx)
	if err != nil {
		return nil, errors.Errorf("查询ElasticSearch失败: %v", err)
	}
	return result, nil
}

func (ec *Client) ScrollQuery(ctx context.Context, r *RequestOption) ([]*elastic.SearchHit, error) {
	query := BuildQuery(r)
	sorts := BuildSort(r)
	fetchSource := BuildFetchSourceContext(r)
	indices := strings.Split(r.Index, ",")
	types := strings.Split(r.Type, ",")
	service := ec.esClient.Scroll().
		SortBy(sorts...).
		Index(indices...).
		Type(types...).
		Size(r.PageSize).
		RestTotalHitsAsInt(true).
		Scroll(r.ScrollKeepAlive).
		FetchSourceContext(fetchSource).
		Query(query)
	var allHits = make([]*elastic.SearchHit, 0)
	result, err := service.Do(ctx)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return allHits, nil
		}
		logs.Errorf("查询ElasticSearch失败: %v", err)
		return nil, err
	}
	allHits = append(allHits, result.Hits.Hits...)
	for {
		result, err = service.Scroll("1m").ScrollId(result.ScrollId).Do(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logs.Errorf("Scroll查询ElasticSearch失败: %v", err)
			return nil, err
		}
		if len(result.Hits.Hits) == 0 {
			break
		} else {
			allHits = append(allHits, result.Hits.Hits...)
		}
	}
	return allHits, nil
}

// Request 请求
func (ec *Client) Request(ctx context.Context, r *RequestOption) (*elastic.SearchResult, error) {
	query := BuildQuery(r)
	sorts := BuildSort(r)
	fetchSource := BuildFetchSourceContext(r)
	indices := strings.Split(r.Index, ",")
	types := strings.Split(r.Type, ",")
	service := ec.esClient.Search().
		SortBy(sorts...).
		Index(indices...).
		Type(types...).
		From(r.PageNum * r.PageSize).
		Size(r.PageSize).
		RestTotalHitsAsInt(true).
		FetchSourceContext(fetchSource).
		TrackTotalHits(true).
		Query(query)
	for _, aggReq := range r.AggReqs {
		BuildAggregate(service, aggReq)
	}
	result, err := service.Do(ctx)
	if err != nil {
		return nil, errors.Errorf("查询ElasticSearch失败: %v", err)
	}
	return result, nil
}
