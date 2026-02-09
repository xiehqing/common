package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/xiehqing/common/agent/agent"
	"github.com/xiehqing/common/agent/app"
	"github.com/xiehqing/common/agent/db"
	"github.com/xiehqing/common/agent/message"
	"github.com/xiehqing/common/agent/session"
	"github.com/xiehqing/common/pkg/logs"
	"gorm.io/gorm"
	"strings"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// handleSession 处理session
func (s *Service) handleSession(ctx context.Context, dbClient *gorm.DB, userId, appId int64, app *app.App, sessionID string) (session.Session, error) {
	session := session.Session{
		ID: sessionID,
	}
	if session.ID == "" {
		newSession, err := app.Sessions.Create(ctx, "")
		if err != nil {
			return session, fmt.Errorf("could not create session:%v", err)
		}
		err = db.CreateAppSession(dbClient, userId, appId, session.ID)
		if err != nil {
			logs.Errorf("could not create app session:%v", err)
		}
		session = newSession
	} else {
		if _, err := app.Sessions.Get(ctx, session.ID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newSession, err := app.Sessions.CreateWithID(ctx, session.ID, "")
				if err != nil {
					return session, fmt.Errorf("could not create session:%v", err)
				}
				session = newSession
				err = db.CreateAppSession(dbClient, userId, appId, session.ID)
				if err != nil {
					logs.Errorf("could not create app session:%v", err)
				}
			} else {
				return session, fmt.Errorf("could not get session:%v", err)
			}
		}
	}
	return session, nil
}

// Run 运行
func (s *Service) Run(ctx context.Context, db *gorm.DB, userId, appId int64, app *app.App, sessionID, prompt string, messageHandler func(message MessageChunk)) (*AgentResponse, error) {
	session, err := s.handleSession(ctx, db, userId, appId, app, sessionID)
	if err != nil {
		return nil, err
	}
	if app.AgentCoordinator == nil {
		return nil, fmt.Errorf("no agent coordinator")
	}
	done := make(chan AgentResponse, 1)
	go func(ctx context.Context, sessionID, prompt string) {
		result, err := app.AgentCoordinator.Run(ctx, session.ID, prompt)
		if err != nil {
			done <- AgentResponse{
				err: fmt.Errorf("failed to start agent processing stream: %w", err),
			}
		}
		done <- AgentResponse{
			result: result,
		}
	}(ctx, session.ID, prompt)
	messageEvents := app.Messages.Subscribe(ctx)
	messageReadBytes := make(map[string]int)
	reasoningMessageReadBytes := make(map[string]int)
	toolCallSend := make(map[string]bool)
	toolResultSend := make(map[string]bool)
	for {
		select {
		case result := <-done:
			if result.err != nil {
				if errors.Is(result.err, context.Canceled) || errors.Is(result.err, agent.ErrRequestCancelled) {
					logs.Infof("Non-interactive: agent processing cancelled，session_id：%s", session.ID)
					return nil, nil
				}
				return nil, fmt.Errorf("agent processing failed: %w", result.err)
			}
			logs.Infof("Non-interactive: agent processing complete，session_id：%s", session.ID)
			return &result, nil
		case event := <-messageEvents:
			msg := event.Payload
			if msg.SessionID == session.ID && msg.Role == message.Tool {
				results := msg.ToolResults()
				if len(results) > 0 {
					for _, tr := range results {
						if toolResultSend[tr.ToolCallID] {
							continue
						}
						// handle tr 防止前端报错
						if tr.Metadata != "" && strings.Contains(tr.Metadata, ".html") {
							tr.Metadata = "{}"
						}
						if messageHandler != nil {
							messageHandler(MessageChunk{
								Role: msg.Role,
								Type: MessageTypeToolResult,
								Data: MessageChunkData{
									ToolResult: &tr,
								},
							})
						}
						toolResultSend[tr.ToolCallID] = true
					}
				}
			}
			if msg.SessionID == session.ID && msg.Role == message.Assistant && len(msg.Parts) > 0 {
				calls := msg.ToolCalls()
				if len(calls) > 0 {
					for _, call := range calls {
						if toolCallSend[call.ID] {
							continue
						}
						if messageHandler != nil {
							messageHandler(MessageChunk{
								Role: msg.Role,
								Type: MessageTypeToolCall,
								Data: MessageChunkData{
									ToolCall: &call,
								},
							})
						}
						toolCallSend[call.ID] = true
					}
				}
				results := msg.ToolResults()
				if len(results) > 0 {
					for _, tr := range results {
						if toolResultSend[tr.ToolCallID] {
							continue
						}
						// handle tr 防止前端报错
						if tr.Metadata != "" && strings.Contains(tr.Metadata, ".html") {
							tr.Metadata = "{}"
						}
						if messageHandler != nil {
							messageHandler(MessageChunk{
								Role: msg.Role,
								Type: MessageTypeToolResult,
								Data: MessageChunkData{
									ToolResult: &tr,
								},
							})
						}
						toolResultSend[tr.ToolCallID] = true
					}
				}
				content := msg.Content().String()
				reasoningContent := msg.ReasoningContent().String()
				readBytes := messageReadBytes[msg.ID]
				reasoningReadBytes := reasoningMessageReadBytes[msg.ID]
				if len(content) < readBytes {
					logs.Errorf("Non-interactive: message content is shorter than read bytes, message_length: %d, read_bytes: %d", len(content), readBytes)
					return nil, fmt.Errorf("message content is shorter than read bytes: %d < %d", len(content), readBytes)
				}
				if len(reasoningContent) < reasoningReadBytes {
					logs.Errorf("Non-interactive: message content is shorter than read bytes, message_length: %d, read_bytes: %d", len(reasoningContent), reasoningReadBytes)
					return nil, fmt.Errorf("message content is shorter than read bytes: %d < %d", len(reasoningContent), reasoningReadBytes)
				}
				reasoningPart := reasoningContent[reasoningReadBytes:]
				if reasoningReadBytes == 0 {
					reasoningPart = strings.TrimLeft(reasoningPart, " \t")
				}
				if len(reasoningPart) > 0 {
					if messageHandler != nil {
						messageHandler(MessageChunk{
							Role: msg.Role,
							Type: MessageTypeReasoning,
							Data: MessageChunkData{
								Thinking: reasoningPart,
							},
						})
					}
				}
				reasoningMessageReadBytes[msg.ID] = len(reasoningContent)
				part := content[readBytes:]
				// Trim leading whitespace. Sometimes the LLM includes leading
				// formatting and intentation, which we don't want here.
				if readBytes == 0 {
					part = strings.TrimLeft(part, " \t")
				}
				if len(part) > 0 {
					if messageHandler != nil {
						messageHandler(MessageChunk{
							Role: msg.Role,
							Type: MessageTypeText,
							Data: MessageChunkData{
								Text: part,
							},
						})
					}
				}
				messageReadBytes[msg.ID] = len(content)
			}

		case <-ctx.Done():
			logs.Infof("Non-interactive: session cancelled，session_id：%s", session.ID)
			return nil, ctx.Err()
		}
	}
}
