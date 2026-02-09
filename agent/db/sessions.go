package db

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionArgs) (Session, error) {
	var s = &Session{
		ParentSessionID:  arg.ParentSessionID.String,
		Title:            arg.Title,
		MessageCount:     arg.MessageCount,
		PromptTokens:     arg.PromptTokens,
		CompletionTokens: arg.CompletionTokens,
		Cost:             arg.Cost,
	}
	s.ID = arg.ID
	err := q.db.Create(s).Error
	if err != nil {
		return *s, err
	}
	return *s, nil
}

func (q *Queries) DeleteSession(ctx context.Context, id string) error {
	err := q.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("session_id = ?", id).Delete(&File{}).Error
		if err != nil {
			return errors.WithMessage(err, "delete files error")
		}
		err = tx.Where("session_id = ?", id).Delete(&Message{}).Error
		if err != nil {
			return errors.WithMessage(err, "delete messages error")
		}
		err = tx.Where("id = ?", id).Delete(&Session{}).Error
		if err != nil {
			return errors.WithMessage(err, "delete session error")
		}
		return nil
	})
	return err
}

func (q *Queries) GetSessionByID(ctx context.Context, id string) (Session, error) {
	var s Session
	err := q.db.Where("id = ?", id).First(&s).Error
	return s, err
}

func (q *Queries) ListSessions(ctx context.Context) ([]Session, error) {
	var sessions []Session
	err := q.db.Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}

func (q *Queries) ListSessionsByIDs(ctx context.Context, ids []string) ([]Session, error) {
	var sessions []Session
	err := q.db.Where("id IN (?)", ids).Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}

func (q *Queries) UpdateSession(ctx context.Context, arg UpdateSessionArgs) (Session, error) {
	session, err := q.GetSessionByID(ctx, arg.ID)
	if err != nil {
		return Session{}, err
	}
	session.Title = arg.Title
	session.PromptTokens = arg.PromptTokens
	session.CompletionTokens = arg.CompletionTokens
	session.SummaryMessageID = arg.SummaryMessageID
	session.Cost = arg.Cost
	session.Todos = arg.Todos
	err = q.db.Save(&session).Error
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

func (q *Queries) UpdateSessionTitleAndUsage(ctx context.Context, arg UpdateSessionTitleAndUsageArgs) error {
	session, err := q.GetSessionByID(ctx, arg.ID)
	if err != nil {
		return err
	}
	session.Title = arg.Title
	session.PromptTokens += arg.PromptTokens
	session.CompletionTokens += arg.CompletionTokens
	session.Cost += arg.Cost
	return q.db.Save(&session).Error
}
