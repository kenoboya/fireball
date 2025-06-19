package repo

import (
	"chat-api/internal/model"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type ChatsRepo struct {
	db *sqlx.DB
}

func NewChatsRepo(db *sqlx.DB) *ChatsRepo {
	return &ChatsRepo{db: db}
}

func (r *ChatsRepo) SetChat(ctx context.Context, chat model.ChatDB) (chatID int64, createdAt time.Time, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, time.Now(), err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chats (creator_id, name, description, type, avatar_url, encrypted)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING chat_id, created_at
	`

	err = tx.QueryRowContext(ctx, query, chat.CreatorID, chat.Name, chat.Description, chat.Type, chat.AvatarURL, chat.Encrypted).Scan(&chatID, &createdAt)
	if err != nil {
		return 0, time.Now(), err
	}

	return chatID, time.Now(), tx.Commit()
}

func (r *ChatsRepo) SetParticipant(ctx context.Context, chatID int64, userID string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO chats_participants (chat_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, chatID, userID)
	return err
}

func (r *ChatsRepo) SetChatRole(ctx context.Context, chatRole model.ChatRole) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chat_roles (chat_id, user_id, granter_id, nickname, role)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = tx.ExecContext(ctx, query, chatRole.ChatID, chatRole.UserID, chatRole.GranterID, chatRole.Nickname, chatRole.Role)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ChatsRepo) SetBlockChat(ctx context.Context, chatID int64, userID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chat_blocked_users (chat_id, user_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ChatsRepo) SetAction(ctx context.Context, chatAction model.ChatAction) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chat_history (chat_id, user_id, action_type)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, query, chatAction.ChatID, chatAction.UserID, chatAction.Type)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ChatsRepo) GetAllChatRoles(ctx context.Context, chatID int64) ([]model.ChatRole, error) {
	var chatRoles []model.ChatRole

	query := `
		SELECT chat_id, user_id, granter_id, nickname, role
		FROM chat_roles
		WHERE chat_id = $1
		ORDER BY role DESC
	`

	err := r.db.SelectContext(ctx, &chatRoles, query, chatID)
	if err != nil {
		return nil, err
	}

	return chatRoles, nil
}

func (r *ChatsRepo) GetAllParticipantsByChatID(ctx context.Context, chatID int64) ([]string, error) {
	var participantsIDs []string

	query := `SELECT user_id FROM chats_participants WHERE chat_id = $1 ORDER BY user_id DESC`

	err := r.db.SelectContext(ctx, &participantsIDs, query, chatID)
	if err != nil {
		return nil, err
	}

	return participantsIDs, nil
}

func (r *ChatsRepo) GetAllActions(ctx context.Context, chatID int64) ([]model.ChatAction, error) {
	var actions []model.ChatAction
	query := `SELECT * FROM chat_history WHERE chat_id = $1 ORDER BY action_timestamp DESC`
	err := r.db.SelectContext(ctx, &actions, query, chatID)
	return actions, err
}

func (r *ChatsRepo) GetAllActionsWithLimit(ctx context.Context, chatID int64, limit int) ([]model.ChatAction, error) {
	var actions []model.ChatAction
	query := `
		SELECT * 
		FROM chat_history 
		WHERE chat_id = $1 
		ORDER BY action_timestamp DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &actions, query, chatID, limit)
	return actions, err
}

func (r *ChatsRepo) GetChatByChatID(ctx context.Context, chatID int64) (model.ChatDB, error) {
	var chat model.ChatDB
	query := `SELECT * FROM chats WHERE chat_id = $1`
	err := r.db.GetContext(ctx, &chat, query, chatID)
	return chat, err
}

func (r *ChatsRepo) GetAllChatsByUserID(ctx context.Context, userID string) ([]model.ChatDB, error) {
	var chats []model.ChatDB
	query := `
		SELECT c.chat_id, c.creator_id, c.name, c.description, c.type, c.created_at, c.updated_at, c.encrypted
		FROM chats c
		JOIN chats_participants cp ON c.chat_id = cp.chat_id
		WHERE cp.user_id = $1
		ORDER BY c.updated_at DESC
	`
	err := r.db.SelectContext(ctx, &chats, query, userID)
	return chats, err
}

func (r *ChatsRepo) GetAllChatsByUserIDWithLimit(ctx context.Context, userID string, limit int) ([]model.ChatDB, error) {
	var chats []model.ChatDB

	query := `
		SELECT c.chat_id, c.creator_id, c.name, c.description, c.type, c.created_at, c.updated_at, c.encrypted
		FROM chats c
		JOIN chats_participants cp ON c.chat_id = cp.chat_id
		WHERE cp.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2
	`

	err := r.db.SelectContext(ctx, &chats, query, userID, limit)
	return chats, err
}

func (r *ChatsRepo) DeleteBlockUser(ctx context.Context, chatID int64, userID string) error {
	query := `DELETE FROM chat_blocked_users WHERE chat_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	return err
}

func (r *ChatsRepo) DeleteChat(ctx context.Context, chatID int64) error {
	query := `DELETE FROM chats WHERE chat_id = $1`
	_, err := r.db.ExecContext(ctx, query, chatID)
	return err
}

func (r *ChatsRepo) IsBlockedChatExists(ctx context.Context, chatID int64, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM chat_blocked_users
			WHERE chat_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, chatID, userID)
	if err != nil {
		return false, err
	}

	return exists, nil
}
