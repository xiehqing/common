package db

import (
	"context"
)

type Querier interface {
	CreateFile(ctx context.Context, arg CreateFileArgs) (File, error)
	CreateMessage(ctx context.Context, arg CreateMessageArgs) (Message, error)
	CreateSession(ctx context.Context, arg CreateSessionArgs) (Session, error)
	DeleteFile(ctx context.Context, id string) error
	DeleteMessage(ctx context.Context, id string) error
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionFiles(ctx context.Context, sessionID string) error
	DeleteSessionMessages(ctx context.Context, sessionID string) error
	GetFile(ctx context.Context, id string) (File, error)
	GetFileByPathAndSession(ctx context.Context, arg GetFileByPathAndSessionArgs) (File, error)
	GetMessage(ctx context.Context, id string) (Message, error)
	GetSessionByID(ctx context.Context, id string) (Session, error)
	ListFilesByPath(ctx context.Context, path string) ([]File, error)
	ListFilesBySession(ctx context.Context, sessionID string) ([]File, error)
	ListLatestSessionFiles(ctx context.Context, sessionID string) ([]File, error)
	ListMessagesBySession(ctx context.Context, sessionID string) ([]Message, error)
	ListNewFiles(ctx context.Context) ([]File, error)
	ListSessions(ctx context.Context) ([]Session, error)
	ListSessionsByIDs(ctx context.Context, ids []string) ([]Session, error)
	UpdateMessage(ctx context.Context, arg UpdateMessageArgs) error
	UpdateSession(ctx context.Context, arg UpdateSessionArgs) (Session, error)
	UpdateSessionTitleAndUsage(ctx context.Context, arg UpdateSessionTitleAndUsageArgs) error
	GetProviders(ctx context.Context) ([]Provider, error)
	GetBigModels(ctx context.Context) ([]BigModel, error)
}

var _ Querier = (*Queries)(nil)
