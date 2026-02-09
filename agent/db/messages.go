package db

import "context"

func (q *Queries) CreateMessage(ctx context.Context, arg CreateMessageArgs) (Message, error) {
	var m = &Message{
		SessionID:        arg.SessionID,
		Role:             arg.Role,
		Parts:            arg.Parts,
		Model:            arg.Model,
		Provider:         arg.Provider,
		IsSummaryMessage: arg.IsSummaryMessage,
	}
	m.ID = arg.ID
	err := q.db.Create(m).Error
	if err != nil {
		return *m, err
	}
	return *m, nil
}

func (q *Queries) DeleteMessage(ctx context.Context, id string) error {
	return q.db.Where("id = ?", id).Delete(&Message{}).Error
}

func (q *Queries) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	return q.db.Where("session_id = ?", sessionID).Delete(&Message{}).Error
}

func (q *Queries) GetMessage(ctx context.Context, id string) (Message, error) {
	var m Message
	err := q.db.Where("id = ?", id).First(&m).Error
	return m, err
}

func (q *Queries) ListMessagesBySession(ctx context.Context, sessionID string) ([]Message, error) {
	var messages []Message
	err := q.db.Where("session_id = ?", sessionID).Order("updated_at ASC").Find(&messages).Error
	return messages, err
}

func (q *Queries) UpdateMessage(ctx context.Context, arg UpdateMessageArgs) error {
	message, err := q.GetMessage(ctx, arg.ID)
	if err != nil {
		return err
	}
	message.Parts = arg.Parts
	message.FinishedAt = arg.FinishedAt
	return q.db.Save(&message).Error
}
