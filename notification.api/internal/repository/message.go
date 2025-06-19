package repo

import (
	"context"
	"notification-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type MessagesRepository struct {
	db *sqlx.DB
}

func NewMessagesRepository(db *sqlx.DB) *MessagesRepository {
	return &MessagesRepository{db: db}
}

func (r *MessagesRepository) SetMessage(ctx context.Context, m model.MessageBriefInfo, action string) (int64, error) {
	formated := m.UpdatedAt.Format("2006-01-02 15:04:05")

	query := `
		INSERT INTO messages (message_external_id, sender_id, type, updated_at, action)
		VALUES (?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query, m.MessageID, m.SenderID, m.Type, formated, action)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MessagesRepository) GetMessage(ctx context.Context, externalMessageID int64) (model.MessageBriefInfo, string, error) {
	var m model.MessageBriefInfo
	var action string
	query := `
		SELECT message_external_id, sender_id, type, updated_at, action
		FROM messages
		WHERE message_external_id = ?
	`
	err := r.db.QueryRowxContext(ctx, query, externalMessageID).Scan(
		&m.MessageID,
		&m.SenderID,
		&m.Type,
		&m.UpdatedAt,
		&action,
	)
	return m, action, err
}

func (r *MessagesRepository) UpdateMessage(ctx context.Context, m model.MessageBriefInfo, action string) error {
	query := `
		UPDATE messages
		SET sender_id = ?, updated_at = ?, action = ?, type = ?
		WHERE message_external_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, m.SenderID, m.UpdatedAt, action, m.Type, m.MessageID)
	return err
}

func (r *MessagesRepository) DeleteMessage(ctx context.Context, externalMessageID int64) error {
	query := `DELETE FROM messages WHERE message_external_id = ?`
	_, err := r.db.ExecContext(ctx, query, externalMessageID)
	return err
}
