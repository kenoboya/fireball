package model

import "time"

const (
	MESSAGE_ACTION_EDITED     = "edited"
	MESSAGE_ACTION_IS_BLURRED = "blurred"
	MESSAGE_ACTION_PASSWORD   = "password"
	MESSAGE_ACTION_REPLIED    = "replied"
	MESSAGE_ACTION_DELETED    = "deleted"
	MESSAGE_ACTION_PINNED     = "pinned"

	CHAT_ACTION_CREATE = "chat was created"
	CHAT_ACTION_DELETE = "chat was deleted"
	CHAT_ACTION_JOIN   = "added to chat by"
	CHAT_ACTION_LEFT   = "left chat"
	CHAT_ACTION_RENAME = "changed the chat name to"
	CHAT_ACTION_KICK   = "was kicked by"

	ACTION_LIMIT_REQUEST = 30
)

type MessageAction struct {
	MessageID int64     `db:"message_id"`
	UserID    string    `db:"user_id"`
	Type      string    `db:"action_type"`
	Time      time.Time `db:"action_timestamp"`
}

type CreateMessageAction struct {
	MessageAction        MessageAction
	CreateMessageRequest CreateMessageRequest
}

type ChatAction struct {
	ChatID int64     `db:"chat_id"`
	UserID string    `db:"user_id"`
	Type   string    `db:"action_type"`
	Time   time.Time `db:"action_timestamp"`
}
