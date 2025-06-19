package repo

import (
	"context"
	"notification-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type UsersRepository struct {
	db *sqlx.DB
}

func NewUsersRepository(db *sqlx.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) SetUser(ctx context.Context, u model.UserBriefInfo) error {
	query := `
		INSERT INTO users (user_id, username, name, avatar_url)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			username = VALUES(username),
			name = VALUES(name),
			avatar_url = VALUES(avatar_url)
	`
	_, err := r.db.ExecContext(ctx, query, u.UserID, u.Username, u.Name, u.AvatarURL)
	return err
}

func (r *UsersRepository) SetUserNotification(ctx context.Context, userNotification model.UserNotification) error {
	query := `
        INSERT INTO user_notifications (user_id, chat_id, mute, term)
        VALUES (:user_id, :chat_id, :mute, :term)
        ON DUPLICATE KEY UPDATE
            chat_id = VALUES(chat_id),
            mute = VALUES(mute),
            term = VALUES(term)
    `
	_, err := r.db.NamedExecContext(ctx, query, userNotification)
	return err
}

func (r *UsersRepository) GetUserNotifications(ctx context.Context, userID string) ([]model.UserNotification, error) {
	var notif []model.UserNotification
	query := `SELECT user_id, chat_id, mute, term FROM user_notifications WHERE user_id = ?`
	err := r.db.GetContext(ctx, &notif, query, userID)
	return notif, err
}

func (r *UsersRepository) UpdateUserNotification(ctx context.Context, userNotification model.UserNotification) error {
	query := `
        UPDATE user_notifications
        SET chat_id = :chat_id,
            mute = :mute,
            term = :term
        WHERE user_id = :user_id
    `
	_, err := r.db.NamedExecContext(ctx, query, userNotification)
	return err
}

func (r *UsersRepository) UpdateUserNotificationForChat(ctx context.Context, userNotification model.UserNotification) error {
	query := `
        UPDATE user_notifications
        SET mute = :mute,
            term = :term
        WHERE user_id = :user_id AND chat_id = :chat_id
    `
	result, err := r.db.NamedExecContext(ctx, query, userNotification)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return model.ErrUserNotificationNotFound
	}

	return nil
}

func (r *UsersRepository) DeleteUserNotification(ctx context.Context, userID string) error {
	query := `DELETE FROM user_notifications WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *UsersRepository) GetUserMutedChat(ctx context.Context, userID string) ([]model.ChatBriefInfo, error) {
	var chats []model.ChatBriefInfo

	query := `
		SELECT 
			c.chat_external_id AS chat_id, 
			c.creator_id, 
			c.name, 
			c.updated_at, 
			c.encrypted, 
			c.avatar_url
		FROM user_notifications un
		JOIN chats c ON c.id = un.chat_id
		WHERE un.user_id = ?
			AND un.mute = TRUE
			AND (un.term IS NULL OR un.term > NOW())
		ORDER BY c.updated_at DESC
	`

	err := r.db.SelectContext(ctx, &chats, query, userID)
	if err != nil {
		return nil, err
	}

	return chats, nil
}
