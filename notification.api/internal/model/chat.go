package model

import "time"

const (
	CHAT_ACTION_CREATE = "chat was created"
	CHAT_ACTION_DELETE = "chat was deleted"
	CHAT_ACTION_JOIN   = "added to chat by"
	CHAT_ACTION_LEFT   = "left chat"
	CHAT_ACTION_RENAME = "changed the chat name to"
	CHAT_ACTION_KICK   = "was kicked by"
)

type ChatBriefInfo struct {
	ChatID    int64     `json:"chat_id" db:"chat_id"`
	CreatorID string    `json:"creator_id" db:"creator_id"`
	Name      string    `json:"name" db:"name"`
	Encrypted bool      `json:"encrypted" db:"encrypted"` // E2EE
	AvatarURL *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
