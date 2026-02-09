package appsessions

import (
	"context"
	"github.com/xiehqing/common/agent/db"
)

type Service interface {
	CreateAppSession(ctx context.Context, userId, appId int64, sessionId string) error
}

type service struct {
	q db.Querier
}

func NewService(q db.Querier) Service {
	return &service{q: q}
}

func (s *service) CreateAppSession(ctx context.Context, userId, appId int64, sessionId string) error {
	return s.q.CreateAppSession(ctx, userId, appId, sessionId)
}
