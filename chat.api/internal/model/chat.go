package model

import "time"

const (
	CHAT_TYPE_PRIVATE = "private"
	CHAT_TYPE_GROUP   = "group"
	CHAT_TYPE_CHANNEL = "channel"

	CHAT_ROLE_USER  = "user"
	CHAT_ROLE_ADMIN = "admin"

	CHAT_LIMIT_REQUEST        = 10
	PINNED_CHAT_LIMIT_REQUEST = 10
)

type ChatDB struct {
	ChatID      int64     `json:"chat_id,omitempty" db:"chat_id"`
	CreatorID   string    `json:"creator_id" db:"creator_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	Type        string    `json:"type" db:"type"`
	Encrypted   bool      `json:"encrypted,omitempty" db:"encrypted"` // E2EE
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" db:"updated_at"`
	AvatarURL   *string   `json:"avatar_url,omitempty" db:"avatar_url"`
}

type ChatBriefInfo struct {
	ChatID    int64     `json:"chat_id" db:"chat_id"`
	CreatorID string    `json:"creator_id" db:"creator_id"`
	Name      string    `json:"name" db:"name"`
	Encrypted bool      `json:"encrypted" db:"encrypted"` // E2EE
	AvatarURL *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ChatRole struct {
	ChatID    int64  `json:"chat_id" db:"chat_id"`
	UserID    string `json:"user_id" db:"user_id"`
	GranterID string `json:"granter_id" db:"granter_id"`
	Nickname  string `json:"nickname" db:"nickname"`
	Role      string `json:"role" db:"role"`
}

type PinnedChat struct {
	ChatID   int64  `json:"chat_id" db:"chat_id"`
	UserID   string `json:"user_id" db:"user_id"`
	Priority int8   `json:"priority" db:"priority"`
}

type CreateGroupChatRequest struct {
	Chat            ChatDB
	ParticipantsIDs []string
	ChatAction      ChatAction
}

type CreatePrivateChatRequest struct {
	Chat           ChatDB               `json:"chat"`
	RecipientID    string               `json:"recipient_id"`
	InitialMessage CreateMessageRequest `json:"initial_message"`
}

type CreatePrivateChatResponse struct {
	Chat        ChatDB
	RecipientID string
	Message     Message
}

type Chat struct {
	ChatDB          ChatDB
	Messages        []Message // last 75 messages
	ParticipantsIDs []string  // if chat active - last 75
	ChatAction      *[]ChatAction
	ChatRoles       *[]ChatRole
}

type PinnedChatInit struct {
	Chat       Chat
	PinnedChat PinnedChat
}

type PinnedChatWithFlag struct {
	PinnedChat PinnedChat `json:"pinned_chat"`
	Fix        bool       `json:"fix"`
}

type BlockChat struct {
	ChatID  int64  `json:"chat_id"`
	UserID  string `json:"user_id"`
	Blocked bool   `json:"blocked"`
}
