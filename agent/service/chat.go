package service

import (
	"charm.land/fantasy"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hatcher/common/agent/agent"
	"github.com/hatcher/common/agent/app"
	"github.com/hatcher/common/agent/message"
	"github.com/hatcher/common/agent/projects"
	"github.com/hatcher/common/agent/session"
	"log/slog"
	"strings"
)

type Service struct {
	app *app.App
}

func NewService(app *app.App) *Service {
	return &Service{app: app}
}

func (s *Service) SendMessage(ctx context.Context, session session.Session, prompt string) (any, error) {
	list, err := projects.List()
	if err != nil {
		return nil, err
	}
	p, _ := json.Marshal(list)
	fmt.Printf("project list:%v", string(p))
	if session.ID == "" {
		newSession, err := s.app.Sessions.Create(context.Background(), "")
		if err != nil {
			return nil, fmt.Errorf("could not create session:%v", err)
		}
		session = newSession
	}
	if s.app.AgentCoordinator == nil {
		return nil, fmt.Errorf("no agent coordinator")
	}
	type response struct {
		result *fantasy.AgentResult
		err    error
	}
	done := make(chan response, 1)

	go func(ctx context.Context, sessionID, prompt string) {
		result, err := s.app.AgentCoordinator.Run(ctx, session.ID, prompt)
		if err != nil {
			done <- response{
				err: fmt.Errorf("failed to start agent processing stream: %w", err),
			}
		}
		done <- response{
			result: result,
		}
	}(ctx, session.ID, prompt)

	messageEvents := s.app.Messages.Subscribe(ctx)
	messageReadBytes := make(map[string]int)
	reasoningMessageReadBytes := make(map[string]int)

	for {
		select {
		case result := <-done:
			//stopSpinner()
			marshal, _ := json.Marshal(result)
			fmt.Printf("\nresult:%s", marshal)
			if result.err != nil {
				if errors.Is(result.err, context.Canceled) || errors.Is(result.err, agent.ErrRequestCancelled) {
					slog.Info("Non-interactive: agent processing cancelled", "session_id", session.ID)
					return nil, nil
				}
				return nil, fmt.Errorf("agent processing failed: %w", result.err)
			}
			return nil, nil

		case event := <-messageEvents:
			msg := event.Payload
			if msg.SessionID == session.ID && msg.Role == message.Assistant && len(msg.Parts) > 0 {

				content := msg.Content().String()
				reasoningContent := msg.ReasoningContent().String()
				calls := msg.ToolCalls()
				if len(calls) > 0 {
					for _, call := range calls {
						fmt.Printf("tool_call:%s, name: %s, input: %s", call.ID, call.Name, call.Input)
					}
				}
				results := msg.ToolResults()
				if len(results) > 0 {
					for _, tr := range results {
						fmt.Printf("tool_result:%s, name: %s, output: %s, content: %s", tr.ToolCallID, tr.Name, tr.Data, tr.Content)
					}
				}

				readBytes := messageReadBytes[msg.ID]
				reasoningReadBytes := reasoningMessageReadBytes[msg.ID]

				if len(content) < readBytes {
					slog.Error("Non-interactive: message content is shorter than read bytes", "message_length", len(content), "read_bytes", readBytes)
					return nil, fmt.Errorf("message content is shorter than read bytes: %d < %d", len(content), readBytes)
				}

				if len(reasoningContent) < reasoningReadBytes {
					slog.Error("Non-interactive: message content is shorter than read bytes", "message_length", len(reasoningContent), "read_bytes", reasoningReadBytes)
					return nil, fmt.Errorf("message content is shorter than read bytes: %d < %d", len(reasoningContent), reasoningReadBytes)
				}
				reasoningPart := reasoningContent[reasoningReadBytes:]
				if reasoningReadBytes == 0 {
					reasoningPart = strings.TrimLeft(reasoningPart, " \t")
				}
				if len(reasoningPart) > 0 {
					fmt.Printf("\nreasoningPart:%s", reasoningPart)
				}
				reasoningMessageReadBytes[msg.ID] = len(reasoningContent)
				part := content[readBytes:]
				// Trim leading whitespace. Sometimes the LLM includes leading
				// formatting and intentation, which we don't want here.
				if readBytes == 0 {
					part = strings.TrimLeft(part, " \t")
				}
				//fmt.Fprint(output, part)
				if len(part) > 0 {
					fmt.Printf("\npart:%s", part)
				}
				messageReadBytes[msg.ID] = len(content)
			}

		case <-ctx.Done():
			fmt.Printf("\n完成")
			return nil, ctx.Err()
		}
	}

	//result, err := s.app.AgentCoordinator.Run(context.Background(), session.ID, prompt, attachments...)
	//if err != nil {
	//	return nil, fmt.Errorf("agent run error: %v", err)
	//}
	//return result, nil
}
