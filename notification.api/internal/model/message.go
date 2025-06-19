package model

import "time"

const (
	MESSAGE_ACTION_SEND       = "send"
	MESSAGE_ACTION_EDITED     = "edited"
	MESSAGE_ACTION_IS_BLURRED = "blurred"
	MESSAGE_ACTION_PASSWORD   = "password"
	MESSAGE_ACTION_REPLIED    = "replied"
	MESSAGE_ACTION_DELETED    = "deleted"
	MESSAGE_ACTION_PINNED     = "pinned"
)

type MessageBriefInfo struct {
	MessageID int64     `json:"message_id" db:"message_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Type      string    `json:"type" db:"type"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
