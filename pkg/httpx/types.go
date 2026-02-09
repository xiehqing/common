package httpx

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"strings"
)

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	DELETE HttpMethod = "DELETE"
	PATCH  HttpMethod = "PATCH"
	HEAD   HttpMethod = "HEAD"
)

func (m HttpMethod) String() string {
	return string(m)
}

type RequestOption struct {
	Method    HttpMethod
	Path      string
	Headers   map[string]string
	Body      interface{}
	Query     map[string]string
	PrintLog  bool
	Sensitive bool
	RequestID string
}

type Option func(option *RequestOption)

func WithMethod(method HttpMethod) Option {
	return func(option *RequestOption) {
		option.Method = method
	}
}

func WithMethodGet() Option {
	return WithMethod(GET)
}

func WithMethodPost() Option {
	return WithMethod(POST)
}

func WithMethodPut() Option {
	return WithMethod(PUT)
}

func WithMethodDelete() Option {
	return WithMethod(DELETE)
}

func WithMethodPatch() Option {
	return WithMethod(PATCH)
}

func WithMethodHead() Option {
	return WithMethod(HEAD)
}

func WithPath(path string) Option {
	return func(option *RequestOption) {
		option.Path = path
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(option *RequestOption) {
		option.Headers = headers
	}
}

func WithBody(body interface{}) Option {
	return func(option *RequestOption) {
		option.Body = body
	}
}

func WithQuery(query map[string]string) Option {
	return func(option *RequestOption) {
		option.Query = query
	}
}

func WithQueryParam(key, value string) Option {
	return func(option *RequestOption) {
		option.Query[key] = value
	}
}

func WithPrintLog(printLog bool) Option {
	return func(option *RequestOption) {
		option.PrintLog = printLog
	}
}

func WithSensitive(sensitive bool) Option {
	return func(option *RequestOption) {
		option.Sensitive = sensitive
	}
}

func NewRequestOption(options ...Option) *RequestOption {
	option := &RequestOption{
		Headers:   make(map[string]string),
		Query:     make(map[string]string),
		PrintLog:  false,
		RequestID: uuid.New().String(),
	}
	for _, opt := range options {
		opt(option)
	}
	return option
}

type RequestLog struct {
	Timestamp string            `json:"timestamp"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      interface{}       `json:"body,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type ResponseLog struct {
	Timestamp  string      `json:"timestamp"`
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	DurationMs int64       `json:"duration_ms,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// LogRequestJSON 请求日志（JSON单行）
func LogRequestJSON(req *RequestLog, isSensitive bool) {
	if isSensitive {
		req.Headers = sensitiveHeaders(req.Headers)
	}
	if jsonStr, err := json.Marshal(req); err == nil {
		log.Printf("HTTP_REQUEST: %s", string(jsonStr))
	}
}

// LogResponseJSON 响应日志（JSON单行）
func LogResponseJSON(resp *ResponseLog) {
	if jsonStr, err := json.Marshal(resp); err == nil {
		log.Printf("HTTP_RESPONSE: %s", string(jsonStr))
	}
}

// 辅助函数
func sensitiveHeaders(headers map[string]string) map[string]string {
	clean := make(map[string]string)
	sensitive := []string{"authorization", "cookie", "token", "password", "secret", "api_secret", "apisecret"}

	for k, v := range headers {
		keyLower := strings.ToLower(k)
		isSensitive := false

		for _, s := range sensitive {
			if strings.Contains(keyLower, s) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			clean[k] = "***REDACTED***"
		} else {
			clean[k] = v
		}
	}

	return clean
}

// BufferedResponse 包装的响应体，允许重复读取
type BufferedResponse struct {
	*http.Response
	bodyBuffer []byte
}

func (br *BufferedResponse) GetBody() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(br.bodyBuffer))
}
