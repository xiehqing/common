package resp

type ResponseCode int64

const (
	Succeeded  ResponseCode = 0
	Failed     ResponseCode = 1
	BadRequest ResponseCode = 400
)

type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data"`
}

func NewResponse(code ResponseCode, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func Success(data interface{}) *Response {
	return NewResponse(Succeeded, "OK", data)
}

func Message(msg string) *Response {
	return NewResponse(Succeeded, msg, nil)
}

func Error(code ResponseCode, message string) *Response {
	return NewResponse(code, message, nil)
}

// PageEntity 分页对象
type PageEntity struct {
	Total    int64 `json:"total"`
	Content  any   `json:"content"`
	PageNo   int   `json:"pageNo"`
	PageSize int   `json:"pageSize"`
}

// NewPageEntity 创建分页对象
func NewPageEntity(total int64, pageNo, pageSize int, data any) PageEntity {
	return PageEntity{
		Total:    total,
		Content:  data,
		PageNo:   pageNo,
		PageSize: pageSize,
	}
}
