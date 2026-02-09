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
	GetAllApps(ctx context.Context, status *int) ([]*Apps, error)
	UpdateApp(ctx context.Context, app *Apps) error
	GetAppSessions(ctx context.Context, appId int64) ([]*AppSessions, error)
	CreateAppSession(ctx context.Context, userId, appId int64, sessionId string) error
	DeleteAppSession(ctx context.Context, sessionId string) error
	GetDataAppSession(ctx context.Context, tenantId, userId, appId int64, dataType, dataId string) (*AppSessions, error)
	CreateDataAppSession(ctx context.Context, tenantId, userId, appId int64, dataType, dataId, sessionId string) error
	GetAppSessionsByDataIds(ctx context.Context, appId int64, dataType string, dataId []string) ([]*AppSessions, error)
	CreateAppUploadFiles(ctx context.Context, files []*AppUploadFiles) error
	UpdateAppUploadFiles(ctx context.Context, files []*AppUploadFiles) error
	GetAppUploadFiles(ctx context.Context, appID int64, userID *int64, sessionID string) ([]*AppUploadFiles, error)
	GetAppSessionUploadFiles(ctx context.Context, sessionID string, targetFiles []string) ([]*AppUploadFiles, error)
}

var _ Querier = (*Queries)(nil)
