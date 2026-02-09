package service

import (
	"charm.land/fantasy"
	"github.com/xiehqing/common/agent/message"
)

type MessageType string

const (
	MessageTypeRequestID  MessageType = "request_id"
	MessageTypeTips       MessageType = "tips"
	MessageTypeReasoning  MessageType = "reasoning"
	MessageTypeToolCall   MessageType = "tool_call"
	MessageTypeToolResult MessageType = "tool_result"
	MessageTypeText       MessageType = "text"
	MessageTypeError      MessageType = "error"
	MessageTypeSessionID  MessageType = "session_id"
)

type AgentResponse struct {
	result *fantasy.AgentResult
	err    error
}

// MessageChunk 消息块
type MessageChunk struct {
	Role message.MessageRole `json:"role"`
	Type MessageType         `json:"type"`
	Data MessageChunkData    `json:"data"`
}

type OriginMessageChunk struct {
	Type MessageType `json:"type"`
	Data any         `json:"data"`
}

func SessionIDChunk(sessionID string) MessageChunk {
	return MessageChunk{
		Type: MessageTypeSessionID,
		Data: MessageChunkData{
			SessionID: sessionID,
		},
	}
}

func RequestIDChunk(requestID string) MessageChunk {
	return MessageChunk{
		Type: MessageTypeRequestID,
		Data: MessageChunkData{
			Text: requestID,
		},
	}
}

func ErrorChunk(error string) MessageChunk {
	return MessageChunk{
		Type: MessageTypeError,
		Data: MessageChunkData{
			Error: error,
		},
	}
}

func TipChunk(tips string) MessageChunk {
	return MessageChunk{
		Type: MessageTypeTips,
		Data: MessageChunkData{
			Tips: tips,
		},
	}
}

// MessageChunkData 消息块数据
type MessageChunkData struct {
	SessionID  string              `json:"session_id,omitempty"`
	ToolCall   *message.ToolCall   `json:"tool_call,omitempty"`
	ToolResult *message.ToolResult `json:"tool_result,omitempty"`
	Thinking   string              `json:"thinking,omitempty"`
	Text       string              `json:"text,omitempty"`
	Tips       string              `json:"tips,omitempty"`
	Error      string              `json:"error,omitempty"`
}
