package repo

import (
	"chat-api/internal/model"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PinnedRepo struct {
	db *sqlx.DB
}

func NewPinnedRepo(db *sqlx.DB) *PinnedRepo {
	return &PinnedRepo{db: db}
}

func (r *PinnedRepo) SetPinnedMessage(ctx context.Context, pinMessage model.PinnedMessage) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pinned_messages (chat_id, message_id, pinned_by_user_id, priority)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, query,
		pinMessage.ChatID,
		pinMessage.MessageID,
		pinMessage.PinnedByUserID,
		pinMessage.Priority,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PinnedRepo) GetPinnedMessagesByChatID(ctx context.Context, chatID int64) ([]model.PinnedMessage, error) {
	var pinnedMessage []model.PinnedMessage

	query := `
		SELECT chat_id, message_id, pinned_by_user_id, priority
		FROM pinned_messages
		WHERE chat_id = $1
		ORDER BY priority ASC
	`
	err := r.db.SelectContext(ctx, &pinnedMessage, query, chatID)
	if err != nil {
		return nil, err
	}

	return pinnedMessage, nil
}

func (r *PinnedRepo) GetPinnedMessageByMessageID(ctx context.Context, messageID int64) (model.PinnedMessage, error) {
	var pinnedMessage model.PinnedMessage

	query := `
		SELECT chat_id, message_id, pinned_by_user_id, priority
		FROM pinned_messages
		WHERE message_id = $1
		ORDER BY priority ASC
	`
	err := r.db.GetContext(ctx, &pinnedMessage, query, messageID)
	if err != nil {
		return model.PinnedMessage{}, err
	}

	return pinnedMessage, nil
}

func (r *PinnedRepo) DeletePinnedMessage(ctx context.Context, pinMessage model.PinnedMessage) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM pinned_messages
		WHERE chat_id = $1 AND message_id = $2
	`

	_, err = tx.ExecContext(ctx, query,
		pinMessage.ChatID,
		pinMessage.MessageID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PinnedRepo) SetPinnedChat(ctx context.Context, pinChat model.PinnedChat) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pinned_chats (chat_id, user_id, priority)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, query,
		pinChat.ChatID,
		pinChat.UserID,
		pinChat.Priority,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PinnedRepo) GetPinnedChats(ctx context.Context, userID string) ([]model.ChatDB, error) {
	var chats []model.ChatDB

	query := `
		SELECT c.*
		FROM chats c
		JOIN pinned_chats p ON c.chat_id = p.chat_id
		WHERE p.user_id = $1
		ORDER BY p.priority ASC
	`

	err := r.db.SelectContext(ctx, &chats, query, userID)
	if err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *PinnedRepo) GetPinnedChatsByUserID(ctx context.Context, userID string) ([]model.PinnedChat, error) {
	var pinnedChats []model.PinnedChat

	query := `
		SELECT chat_id, user_id, priority
		FROM pinned_chats
		WHERE user_id = $1
		ORDER BY priority ASC
	`
	err := r.db.SelectContext(ctx, &pinnedChats, query, userID)
	if err != nil {
		return nil, err
	}

	return pinnedChats, nil
}

func (r *PinnedRepo) GetPinnedChatsByUserIDWithLimit(ctx context.Context, userID string, limit int) ([]model.PinnedChat, error) {
	var pinnedChats []model.PinnedChat

	query := `
		SELECT chat_id, user_id, priority
		FROM pinned_chats
		WHERE user_id = $1
		ORDER BY priority ASC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &pinnedChats, query, userID, limit)
	if err != nil {
		return nil, err
	}

	return pinnedChats, nil
}

func (r *PinnedRepo) UpdatePinnedChat(ctx context.Context, pinChat model.PinnedChat) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE pinned_chats
		SET priority = $3
		WHERE chat_id = $1 AND user_id = $2
	`

	res, err := tx.ExecContext(ctx, query,
		pinChat.ChatID,
		pinChat.UserID,
		pinChat.Priority,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pinned chat not found for update: chat_id=%v user_id=%s", pinChat.ChatID, pinChat.UserID)
	}

	return tx.Commit()
}

func (r *PinnedRepo) IsPinnedChatExists(ctx context.Context, pinChat model.PinnedChat) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pinned_chats WHERE chat_id = $1 AND user_id = $2
		)
	`

	err := r.db.GetContext(ctx, &exists, query, pinChat.ChatID, pinChat.UserID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PinnedRepo) DeletePinnedChat(ctx context.Context, pinChat model.PinnedChat) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM pinned_chats
		WHERE chat_id = $1 AND user_id = $2
	`

	_, err = tx.ExecContext(ctx, query,
		pinChat.ChatID,
		pinChat.UserID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
