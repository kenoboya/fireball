package repo

import (
	"chat-api/internal/model"
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type MessagesRepo struct {
	db *sqlx.DB
}

func NewMessagesRepo(db *sqlx.DB) *MessagesRepo {
	return &MessagesRepo{db: db}
}

func (r *MessagesRepo) SetMessage(ctx context.Context, message model.MessageDB) (messageID int64, createdAt time.Time, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO messages (sender_id, content, status, type)
		VALUES ($1, $2, $3, $4)
		RETURNING message_id, created_at
	`

	err = tx.QueryRowContext(ctx, query, message.SenderID, message.Content, message.Status, message.Type).
		Scan(&messageID, &createdAt)
	if err != nil {
		return 0, time.Time{}, err
	}

	return messageID, createdAt, tx.Commit()
}

func (r *MessagesRepo) SetAction(ctx context.Context, messageAction model.MessageAction) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO message_audio_log (message_id, user_id, action_type)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, query, messageAction.MessageID, messageAction.UserID, messageAction.Type)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessagesRepo) GetMessageByMessageID(ctx context.Context, messageID int64) (model.MessageDB, error) {
	var message model.MessageDB

	query := `
		SELECT message_id, sender_id, content, status, type, created_at, updated_at
		FROM messages
		WHERE message_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, messageID).
		Scan(&message.MessageID, &message.SenderID, &message.Content, &message.Status, &message.Type, &message.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return message, nil
		}
		return message, err
	}

	return message, nil
}

func (r *MessagesRepo) GetMessagesByChatID(ctx context.Context, chatID int64) ([]model.MessageDB, error) {
	var messages []model.MessageDB

	query := `
		SELECT m.message_id, m.sender_id, m.content, m.status, m.type, m.created_at, m.updated_at
		FROM messages m
		JOIN chat_messages cm ON m.message_id = cm.message_id
		WHERE cm.chat_id = $1
		ORDER BY m.created_at DESC
	`

	err := r.db.SelectContext(ctx, &messages, query, chatID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessagesRepo) GetMessagesByChatIDWithLimit(ctx context.Context, chatID int64, limit int) ([]model.MessageDB, error) {
	var messages []model.MessageDB

	query := `
		SELECT m.message_id, m.sender_id, m.content, m.status, m.type, m.created_at, m.updated_at
		FROM messages m
		JOIN chat_messages cm ON m.message_id = cm.message_id
		WHERE cm.chat_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`

	err := r.db.SelectContext(ctx, &messages, query, chatID, limit)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessagesRepo) GetAllActions(ctx context.Context, messageID int64) ([]model.MessageAction, error) {
	var actions []model.MessageAction

	query := `
		SELECT *
		FROM message_audit_log
		WHERE message_id = $1
		ORDER BY action_timestamp DESC
	`

	err := r.db.SelectContext(ctx, &actions, query, messageID)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

func (r *MessagesRepo) DeleteMessage(ctx context.Context, messageID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessagesRepo) SetBindMessageMedia(ctx context.Context, messageID, mediaID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO messages_media (message_id, media_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, messageID, mediaID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessagesRepo) SetBindMessageLocation(ctx context.Context, messageID, locationID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO messages_locations (message_id, location_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, messageID, locationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessagesRepo) SetBindMessageFile(ctx context.Context, messageID, fileID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO messages_files (message_id, file_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, messageID, fileID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessagesRepo) SetBindMessageChat(ctx context.Context, messageID, chatID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chat_messages (chat_id, message_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, chatID, messageID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
