package model

import "time"

const (
	MESSAGE_LIMIT_REQUEST = 75
	MESSAGE_SENT          = "sent"
	MESSAGE_DELIVERED     = "delivered"
	MESSAGE_READ          = "read"

	MESSAGE_TEXT     = "text"
	MESSAGE_MEDIA    = "media"
	MESSAGE_FILE     = "file"
	MESSAGE_LOCATION = "location"
	MESSAGE_MIXED    = "mixed"

	VERY_HIGH = iota + 1 // 1
	HIGH                 // 2
	MEDIUM               // 3
	LOW                  // 4
	VERY_LOW             // 5
)

type MessageDB struct {
	MessageID int64     `json:"message_id,omitempty" db:"message_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Content   *string   `json:"content,omitempty" db:"content"`
	Status    string    `json:"status" db:"status"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MessageBriefInfo struct {
	MessageID int64     `json:"message_id" db:"message_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Type      string    `json:"type" db:"type"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	MessageWithData SendMessage
	PinnedMessage   *PinnedMessage
	Action          *[]MessageAction
}

type SendMessage struct {
	MessageDB MessageDB   `json:"message"`
	Media     *[]Media    `json:"media"`
	Locations *[]Location `json:"locations"`
	Files     *[]File     `json:"files"`
}

type PinnedMessage struct {
	ChatID         int64  `json:"chat_id" db:"chat_id"`
	MessageID      int64  `json:"message_id" db:"message_id"`
	PinnedByUserID string `json:"pinned_by_user_id" db:"pinned_by_user_id"`
	Priority       *int8  `json:"priority" db:"priority"`
}

type CreateMessageRequest struct {
	ChatID          int64       `json:"chat_id,omitempty"`
	MessageWithData SendMessage `json:"first_message"`
}
