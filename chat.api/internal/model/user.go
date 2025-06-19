package model

const (
	USER_LIMIT_REQUEST = 200
)

type User struct {
	UserID      string  `json:"user_id" db:"_id"`
	Username    string  `json:"username" db:"username"`
	DisplayName string  `json:"display_name" db:"display_name"`
	Bio         *string `json:"bio" db:"bio"`
	Email       *string `json:"email,omitempty" db:"email"`
	Phone       *string `json:"phone,omitempty" db:"phone"`
	AvatarURL   *string `bson:"avatar_url"`
}

type UserBriefInfo struct {
	UserID    string  `json:"user_id" db:"user_id"`
	Username  string  `json:"username" db:"username"`
	Name      string  `json:"name" db:"name"` // alias
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}
