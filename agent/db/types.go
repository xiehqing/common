package db

import "database/sql"

type CreateFileArgs struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Version   int64  `json:"version"`
}

type GetFileByPathAndSessionArgs struct {
	Path      string `json:"path"`
	SessionID string `json:"session_id"`
}

type CreateSessionArgs struct {
	ID               string         `json:"id"`
	ParentSessionID  sql.NullString `json:"parent_session_id"`
	Title            string         `json:"title"`
	MessageCount     int64          `json:"message_count"`
	PromptTokens     int64          `json:"prompt_tokens"`
	CompletionTokens int64          `json:"completion_tokens"`
	Cost             float64        `json:"cost"`
}

type UpdateSessionArgs struct {
	Title            string  `json:"title"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	SummaryMessageID string  `json:"summary_message_id"`
	Cost             float64 `json:"cost"`
	Todos            string  `json:"todos"`
	ID               string  `json:"id"`
}

type UpdateSessionTitleAndUsageArgs struct {
	Title            string  `json:"title"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	Cost             float64 `json:"cost"`
	ID               string  `json:"id"`
}

type CreateMessageArgs struct {
	ID               string `json:"id"`
	SessionID        string `json:"session_id"`
	Role             string `json:"role"`
	Parts            string `json:"parts"`
	Model            string `json:"model"`
	Provider         string `json:"provider"`
	IsSummaryMessage int64  `json:"is_summary_message"`
}

type UpdateMessageArgs struct {
	Parts      string `json:"parts"`
	FinishedAt int64  `json:"finished_at"`
	ID         string `json:"id"`
}
