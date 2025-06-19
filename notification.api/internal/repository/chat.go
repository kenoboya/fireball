package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"notification-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type ChatsRepository struct {
	db *sqlx.DB
}

func NewChatsRepository(db *sqlx.DB) *ChatsRepository {
	return &ChatsRepository{db: db}
}

func (r *ChatsRepository) AddMessageToChat(ctx context.Context, internalChatID, internalMessageID int64) error {
	query := `
        INSERT INTO chat_messages (chat_id, message_id)
        VALUES (?, ?)
        ON DUPLICATE KEY UPDATE chat_id = chat_id
    `
	_, err := r.db.ExecContext(ctx, query, internalChatID, internalMessageID)
	return err
}

func (r *ChatsRepository) SetChat(ctx context.Context, c model.ChatBriefInfo, action *string) (int64, error) {
	var (
		query string
		args  []interface{}
	)

	formatted := c.UpdatedAt.Format("2006-01-02 15:04:05")

	if action != nil {
		query = `
			INSERT INTO chats (chat_external_id, creator_id, name, updated_at, encrypted, avatar_url, action)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				updated_at = VALUES(updated_at),
				encrypted = VALUES(encrypted),
				avatar_url = VALUES(avatar_url),
				action = VALUES(action)
		`
		args = []interface{}{c.ChatID, c.CreatorID, c.Name, formatted, c.Encrypted, c.AvatarURL, *action}
	} else {
		query = `
			INSERT INTO chats (chat_external_id, creator_id, name, updated_at, encrypted, avatar_url)
			VALUES (?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				updated_at = VALUES(updated_at),
				encrypted = VALUES(encrypted),
				avatar_url = VALUES(avatar_url)
		`
		args = []interface{}{c.ChatID, c.CreatorID, c.Name, formatted, c.Encrypted, c.AvatarURL}
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert/update chat: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if id == 0 {
		var existingID int64
		err = r.db.GetContext(ctx, &existingID, `
			SELECT id FROM chats WHERE chat_external_id = ? AND creator_id = ?`, c.ChatID, c.CreatorID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("chat not found after insert/update: external_id=%d creator_id=%s", c.ChatID, c.CreatorID)
			}
			return 0, fmt.Errorf("failed to query existing chat: %w", err)
		}
		id = existingID
	}

	return id, nil
}

func (r *ChatsRepository) GetChat(ctx context.Context, externalChatID int64) (model.ChatBriefInfo, string, error) {
	var c model.ChatBriefInfo
	var action string
	query := `
		SELECT chat_external_id, creator_id, name, updated_at, encrypted, avatar_url, action
		FROM chats
		WHERE chat_external_id = ?
	`
	err := r.db.QueryRowxContext(ctx, query, externalChatID).Scan(
		&c.ChatID,
		&c.CreatorID,
		&c.Name,
		&c.UpdatedAt,
		&c.Encrypted,
		&c.AvatarURL,
		&action,
	)
	return c, action, err
}

func (r *ChatsRepository) UpdateChat(ctx context.Context, c model.ChatBriefInfo, action *string) error {
	query := `
		UPDATE chats
		SET creator_id = ?, name = ?, updated_at = ?, action = ?, avatar_url = ?
		WHERE chat_external_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, c.CreatorID, c.Name, c.UpdatedAt, action, c.AvatarURL, c.ChatID)
	return err
}

func (r *ChatsRepository) DeleteChat(ctx context.Context, externalChatID int64) error {
	query := `DELETE FROM chats WHERE chat_external_id = ?`
	_, err := r.db.ExecContext(ctx, query, externalChatID)
	return err
}
