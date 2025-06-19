package repo

import (
	"context"
	"notification-api/internal/model"
	"notification-api/pkg/logger"
	"time"

	"github.com/jmoiron/sqlx"
)

type NotificationsRepository struct {
	db *sqlx.DB
}

func NewNotificationsRepository(db *sqlx.DB) *NotificationsRepository {
	return &NotificationsRepository{db: db}
}

// DB struct for notification messages
type notificationMessageDB struct {
	ChatID        int64   `db:"chat_chat_id"`
	ChatCreatorID string  `db:"chat_creator_id"`
	ChatName      string  `db:"chat_name"`
	ChatEncrypted bool    `db:"chat_encrypted"`
	ChatAvatarURL *string `db:"chat_avatar_url"`
	ChatUpdatedAt string  `db:"chat_updated_at"`

	MessageID      int64  `db:"message_message_id"`
	MessageSender  string `db:"message_sender_id"`
	MessageType    string `db:"message_type"`
	MessageUpdated string `db:"message_updated_at"`
	MessageAction  string `db:"message_action"`

	SenderUserID   string  `db:"sender_user_id"`
	SenderUsername string  `db:"sender_username"`
	SenderName     string  `db:"sender_name"`
	SenderAvatar   *string `db:"sender_avatar_url"`

	RecipientID string `db:"recipient_id"`
	IsRead      bool   `db:"is_read"`
}

func toNotificationMessage(dbRow notificationMessageDB) model.NotificationMessage {
	t_chat, err := time.Parse("2006-01-02 15:04:05", dbRow.ChatUpdatedAt)
	if err != nil {
		logger.Warn("Error with converting time db to time.Time")
	}
	t_message, err := time.Parse("2006-01-02 15:04:05", dbRow.MessageUpdated)
	if err != nil {
		logger.Warn("Error with converting time db to time.Time")
	}
	return model.NotificationMessage{
		Chat: model.ChatBriefInfo{
			ChatID:    dbRow.ChatID,
			CreatorID: dbRow.ChatCreatorID,
			Name:      dbRow.ChatName,
			Encrypted: dbRow.ChatEncrypted,
			AvatarURL: dbRow.ChatAvatarURL,
			UpdatedAt: t_chat,
		},
		Message: model.MessageBriefInfo{
			MessageID: dbRow.MessageID,
			SenderID:  dbRow.MessageSender,
			Type:      dbRow.MessageType,
			UpdatedAt: t_message,
		},
		Sender: model.UserBriefInfo{
			UserID:    dbRow.SenderUserID,
			Username:  dbRow.SenderUsername,
			Name:      dbRow.SenderName,
			AvatarURL: dbRow.SenderAvatar,
		},
		RecipientID:   dbRow.RecipientID,
		MessageAction: dbRow.MessageAction,
	}
}

func (r *NotificationsRepository) SetMessage(ctx context.Context, internalMessageID int64, recipientID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	query := `
		INSERT INTO message_notifications (message_id, recipient_id, is_read)
		VALUES (?, ?, false)
		ON DUPLICATE KEY UPDATE message_id = message_id
	`
	_, err := r.db.ExecContext(ctx, query, internalMessageID, recipientID)
	return err
}

// GetMessage returns a notification message by ID and recipient, only if unread, marks as read
func (r *NotificationsRepository) GetMessage(ctx context.Context, internalMessageID int64, recipientID string) (model.NotificationMessage, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id   AS chat_creator_id,
		c.name         AS chat_name,
		c.encrypted    AS chat_encrypted,
		c.avatar_url   AS chat_avatar_url,
		c.updated_at   AS chat_updated_at,

		m.message_external_id   AS message_message_id,
		m.sender_id    AS message_sender_id,
		m.type         AS message_type,
		m.updated_at   AS message_updated_at,
		m.action       AS message_action,

		u.user_id      AS sender_user_id,
		u.username     AS sender_username,
		u.name         AS sender_name,
		u.avatar_url   AS sender_avatar_url,

		mn.recipient_id AS recipient_id,
		mn.is_read      AS is_read,

	FROM message_notifications mn
	JOIN messages m ON mn.message_id = m.id
	JOIN chat_messages cm ON cm.message_id = m.id
	JOIN chats c ON cm.chat_id = c.id
	JOIN users u ON m.sender_id = u.user_id
	-- ADD: join with user_notifications and filter by mute and term
	LEFT JOIN user_notifications un ON un.user_id = mn.recipient_id AND un.chat_id = c.id
	WHERE mn.message_id = ? AND mn.recipient_id = ? AND mn.is_read = FALSE
	  AND (
	      un.mute = FALSE
	      OR un.mute IS NULL
	      OR (un.term IS NOT NULL AND un.term <= NOW()) -- ADD: expired mute = not muted
	  )
	`

	var dbRow notificationMessageDB
	err := r.db.GetContext(ctx, &dbRow, query, internalMessageID, recipientID)
	if err != nil {
		return model.NotificationMessage{}, err
	}

	select {
	case <-ctx.Done():
		return model.NotificationMessage{}, ctx.Err()
	default:
	}

	updateQuery := `
		UPDATE message_notifications
		SET is_read = TRUE
		WHERE recipient_id = ? AND message_id = ?
	`
	_, err = r.db.ExecContext(ctx, updateQuery, recipientID, internalMessageID)
	if err != nil {
		return model.NotificationMessage{}, err
	}

	return toNotificationMessage(dbRow), nil
}

// GetMessagesForParticipant fetches all unread message notifications for a recipient and marks them read
func (r *NotificationsRepository) GetMessagesForRecipient(ctx context.Context, recipientID string) ([]model.NotificationMessage, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id   AS chat_creator_id,
		c.name         AS chat_name,
		c.encrypted    AS chat_encrypted,
		c.avatar_url   AS chat_avatar_url,
		c.updated_at   AS chat_updated_at,

		m.message_external_id   AS message_message_id,
		m.sender_id    AS message_sender_id,
		m.type         AS message_type,
		m.updated_at   AS message_updated_at,
		m.action       AS message_action,

		u.user_id      AS sender_user_id,
		u.username     AS sender_username,
		u.name         AS sender_name,
		u.avatar_url   AS sender_avatar_url,

		mn.recipient_id AS recipient_id,
		mn.is_read      AS is_read

	FROM message_notifications mn
	JOIN messages m ON mn.message_id = m.id
	JOIN chat_messages cm ON cm.message_id = m.id
	JOIN chats c ON cm.chat_id = c.id
	JOIN users u ON m.sender_id = u.user_id
	-- ADD: join with user_notifications and filter by mute and term
	LEFT JOIN user_notifications un ON un.user_id = mn.recipient_id AND un.chat_id = c.id
	WHERE mn.recipient_id = ? AND mn.is_read = FALSE
	  AND (
	      un.mute = FALSE
	      OR un.mute IS NULL
	      OR (un.term IS NOT NULL AND un.term <= NOW()) -- ADD: expired mute = not muted
	  )
	ORDER BY m.updated_at DESC
	`

	var dbRows []notificationMessageDB
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

	messageIDs := make([]int64, 0, len(dbRows))
	for _, row := range dbRows {
		messageIDs = append(messageIDs, row.MessageID)
	}

	updateQuery := `
		UPDATE message_notifications mn
        JOIN messages m ON mn.message_id = m.id
        SET mn.is_read = TRUE
        WHERE mn.recipient_id = ? AND m.message_external_id IN (?)
	`
	queryWithArgs, args, err := sqlx.In(updateQuery, recipientID, messageIDs)
	if err != nil {
		return nil, err
	}
	queryWithArgs = r.db.Rebind(queryWithArgs)

	_, err = r.db.ExecContext(ctx, queryWithArgs, args...)
	if err != nil {
		return nil, err
	}

	results := make([]model.NotificationMessage, 0, len(dbRows))
	for _, dbRow := range dbRows {
		results = append(results, toNotificationMessage(dbRow))
	}

	return results, nil
}

func (r *NotificationsRepository) GetMessages(ctx context.Context, internalMessageID int64) ([]model.NotificationMessage, error) {
	query := `
	SELECT
		c.chat_external_id  AS chat_chat_id,
		c.creator_id   AS chat_creator_id,
		c.name         AS chat_name,
		c.encrypted    AS chat_encrypted,
		c.avatar_url   AS chat_avatar_url,
		c.updated_at   AS chat_updated_at,

		m.message_external_id   AS message_message_id,
		m.sender_id    AS message_sender_id,
		m.type         AS message_type,
		m.updated_at   AS message_updated_at,
		m.action       AS message_action,

		u.user_id      AS sender_user_id,
		u.username     AS sender_username,
		u.name         AS sender_name,
		u.avatar_url   AS sender_avatar_url,

		mn.recipient_id   AS recipient_id,
		mn.is_read        AS is_read

	FROM message_notifications mn
	JOIN messages m ON mn.message_id = m.id
	JOIN chat_messages cm ON cm.message_id = m.id
	JOIN chats c ON cm.chat_id = c.id
	JOIN users u ON m.sender_id = u.user_id
	WHERE mn.message_id = ?
	`

	var dbRows []notificationMessageDB
	err := r.db.SelectContext(ctx, &dbRows, query, internalMessageID)
	if err != nil {
		return nil, err
	}

	results := make([]model.NotificationMessage, 0, len(dbRows))
	for _, dbRow := range dbRows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		results = append(results, toNotificationMessage(dbRow))
	}

	return results, nil
}
