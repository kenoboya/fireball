package model

import "time"

const (
	MUTED   = true
	UNMUTED = false
)

type UserBriefInfo struct {
	UserID    string  `json:"user_id" db:"user_id"`
	Username  string  `json:"username" db:"username"`
	Name      string  `json:"name" db:"name"` // alias
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

type UserNotification struct {
	UserID string     `json:"user_id" db:"user_id"`
	ChatID string     `json:"chat_id" db:"chat_id"`
	Muted  bool       `json:"muted" db:"muted"`
	Term   *time.Time `json:"term,omitempty" db:"term"`
}
