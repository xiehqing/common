package hertzx

import (
	"github.com/hatcher/common/pkg/util"
	"github.com/hertz-contrib/sse"
)

type SseSender struct {
	ss *sse.Stream
}

func NewSseSender(ss *sse.Stream) *SseSender {
	return &SseSender{ss: ss}
}

// Send 发送
func (s *SseSender) Send(data *sse.Event) error {
	return s.ss.Publish(data)
}

// BuildDataEvent 构建事件
func BuildDataEvent(data any) *sse.Event {
	if data == nil {
		return nil
	}
	if _, ok := data.(*sse.Event); ok {
		return data.(*sse.Event)
	}
	if _, ok := data.(string); ok {
		return &sse.Event{
			Data: []byte(data.(string)),
		}
	}
	m := util.ToJsonIgnoreError(data)
	return &sse.Event{
		Data: []byte(m),
	}
}
