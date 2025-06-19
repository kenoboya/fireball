package repo

import (
	"context"
	"notification-api/internal/model"
	"notification-api/pkg/logger"
	"time"

	"github.com/jmoiron/sqlx"
)

// DB struct for notification chats
type notificationChatDB struct {
	ChatID        int64   `db:"chat_chat_id"`
	ChatCreatorID string  `db:"chat_creator_id"`
	ChatName      string  `db:"chat_name"`
	ChatEncrypted bool    `db:"chat_encrypted"`
	ChatAvatarURL *string `db:"chat_avatar_url"`
	ChatUpdatedAt string  `db:"chat_updated_at"`
	ChatAction    string  `db:"chat_action"`

	SenderUserID   string  `db:"sender_user_id"`
	SenderUsername string  `db:"sender_username"`
	SenderName     string  `db:"sender_name"`
	SenderAvatar   *string `db:"sender_avatar_url"`

	RecipientID string `db:"recipient_id"`
	IsRead      bool   `db:"is_read"`
}

func toNotificationChat(dbRow notificationChatDB) model.NotificationChat {
	t, err := time.Parse("2006-01-02 15:04:05", dbRow.ChatUpdatedAt)
	if err != nil {
		logger.Warn("Error with converting time db to time.Time")
	}
	return model.NotificationChat{
		Chat: model.ChatBriefInfo{
			ChatID:    dbRow.ChatID,
			CreatorID: dbRow.ChatCreatorID,
			Name:      dbRow.ChatName,
			Encrypted: dbRow.ChatEncrypted,
			AvatarURL: dbRow.ChatAvatarURL,
			UpdatedAt: t,
		},
		Sender: model.UserBriefInfo{
			UserID:    dbRow.SenderUserID,
			Username:  dbRow.SenderUsername,
			Name:      dbRow.SenderName,
			AvatarURL: dbRow.SenderAvatar,
		},
		RecipientID: dbRow.RecipientID,
		ChatAction:  dbRow.ChatAction,
	}
}

func (r *NotificationsRepository) SetChat(ctx context.Context, internalChatID int64, recipientID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	query := `
		INSERT INTO chat_notifications (chat_id, recipient_id, is_read)
		VALUES (?, ?, false)
		ON DUPLICATE KEY UPDATE chat_id = chat_id
	`
	_, err := r.db.ExecContext(ctx, query, internalChatID, recipientID)
	return err
}

// GetChat returns a notification chat by ID and recipient, only if unread, marks it as read
func (r *NotificationsRepository) GetChat(ctx context.Context, internalChatID int64, recipientID string) (model.NotificationChat, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id   AS chat_creator_id,
		c.name         AS chat_name,
		c.encrypted    AS chat_encrypted,
		c.avatar_url   AS chat_avatar_url,
		c.updated_at   AS chat_updated_at,
		c.action       AS chat_action,

		u.user_id      AS sender_user_id,
		u.username     AS sender_username,
		u.name         AS sender_name,
		u.avatar_url   AS sender_avatar_url,

		cn.recipient_id AS recipient_id,
		cn.is_read      AS is_read,

	FROM chat_notifications cn n    
	JOIN chats c ON cn.chat_id = c.id
	JOIN users u ON c.creator_id = u.user_id
	-- ADD: join with user_notifications to check mute and term
	LEFT JOIN user_notifications un ON un.user_id = cn.recipient_id AND un.chat_id = c.id
	WHERE cn.chat_id = ? AND cn.recipient_id = ? AND cn.is_read = FALSE
	  AND (
	      un.mute = FALSE 
	      OR un.mute IS NULL
	      OR (un.term IS NOT NULL AND un.term <= NOW()) -- ADD: expired mute = not muted
	  )
	`

	var dbRow notificationChatDB
	err := r.db.GetContext(ctx, &dbRow, query, internalChatID, recipientID)
	if err != nil {
		return model.NotificationChat{}, err
	}

	select {
	case <-ctx.Done():
		return model.NotificationChat{}, ctx.Err()
	default:
	}

	updateQuery := `
		UPDATE chat_notifications
		SET is_read = TRUE
		WHERE recipient_id = ? AND chat_id = ?
	`
	_, err = r.db.ExecContext(ctx, updateQuery, recipientID, internalChatID)
	if err != nil {
		return model.NotificationChat{}, err
	}

	return toNotificationChat(dbRow), nil
}

// GetChatsForParticipant fetches all unread chat notifications for a recipient and marks them read
func (r *NotificationsRepository) GetChatsForRecipient(ctx context.Context, recipientID string) ([]model.NotificationChat, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id   AS chat_creator_id,
		c.name         AS chat_name,
		c.encrypted    AS chat_encrypted,
		c.avatar_url   AS chat_avatar_url,
		c.updated_at   AS chat_updated_at,
		c.action       AS chat_action,

		u.user_id      AS sender_user_id,
		u.username     AS sender_username,
		u.name         AS sender_name,
		u.avatar_url   AS sender_avatar_url,

		cn.recipient_id AS recipient_id,
		cn.is_read      AS is_read

	FROM chat_notifications cn
	JOIN chats c ON cn.chat_id = c.id
	JOIN users u ON c.creator_id = u.user_id
	-- ADD: join with user_notifications to check mute and term
	LEFT JOIN user_notifications un ON un.user_id = cn.recipient_id AND un.chat_id = c.id
	WHERE cn.recipient_id = ? AND cn.is_read = FALSE
	  AND (
	      un.mute = FALSE 
	      OR un.mute IS NULL
	      OR (un.term IS NOT NULL AND un.term <= NOW()) -- ADD: expired mute = not muted
	  )
	ORDER BY c.updated_at DESC
	`

	var dbRows []notificationChatDB
	err := r.db.SelectContext(ctx, &dbRows, query, recipientID)
	if err != nil {
		return nil, err
	}

	if len(dbRows) == 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	chatIDs := make([]int64, 0, len(dbRows))
	for _, row := range dbRows {
		chatIDs = append(chatIDs, row.ChatID)
	}

	// Mark all fetched chats as read
	updateQuery := `
    UPDATE chat_notifications cn
    JOIN chats c ON cn.chat_id = c.id
    SET cn.is_read = TRUE
    WHERE cn.recipient_id = ? AND c.chat_external_id IN (?)
	`

	queryWithArgs, args, err := sqlx.In(updateQuery, recipientID, chatIDs)
	if err != nil {
		return nil, err
	}
	queryWithArgs = r.db.Rebind(queryWithArgs)
	_, err = r.db.ExecContext(ctx, queryWithArgs, args...)
	if err != nil {
		return nil, err
	}

	results := make([]model.NotificationChat, 0, len(dbRows))
	for _, dbRow := range dbRows {
		results = append(results, toNotificationChat(dbRow))
	}

	return results, nil
}

func (r *NotificationsRepository) GetChats(ctx context.Context, internalChatID int64) ([]model.NotificationChat, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id    AS chat_creator_id,
		c.name          AS chat_name,
		c.encrypted     AS chat_encrypted,
		c.avatar_url    AS chat_avatar_url,
		c.updated_at    AS chat_updated_at,
		c.action        AS chat_action,

		u.user_id       AS sender_user_id,
		u.username      AS sender_username,
		u.name          AS sender_name,
		u.avatar_url    AS sender_avatar_url,

		cn.recipient_id AS recipient_id,
		cn.is_read      AS is_read

	FROM chat_notifications cn
	JOIN chats c ON cn.chat_id = c.id
	JOIN users u ON c.creator_id = u.user_id
	WHERE cn.chat_id = ?
	`

	var dbRows []notificationChatDB
	err := r.db.SelectContext(ctx, &dbRows, query, internalChatID)
	if err != nil {
		return nil, err
	}

	results := make([]model.NotificationChat, 0, len(dbRows))
	for _, dbRow := range dbRows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		results = append(results, toNotificationChat(dbRow))
	}

	return results, nil
}
