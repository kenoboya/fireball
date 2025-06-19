package model

type User struct {
	UserID      string  `json:"user_id" bson:"_id"`
	Username    string  `json:"username" bson:"username"`
	DisplayName string  `json:"display_name" bson:"display_name"`
	Bio         *string `json:"bio" bson:"bio"`
	Email       *string `json:"email,omitempty" bson:"email"`
	Phone       *string `json:"phone,omitempty" bson:"phone"`
	AvatarURL   *string `json:"avatar_url,omitempty" bson:"avatar_url"`
}

type UserRequest struct {
	SenderID    string `bson:"sender_id" json:"sender_id"`
	RecipientID string `bson:"recipient_id" json:"recipient_id"`
}

type UserBriefInfo struct {
	UserID    string  `bson:"_id"`
	Username  string  `bson:"username"`
	Name      string  `bson:"display_name"` // alias or display_name
	AvatarURL *string `bson:"avatar_url"`
}

type UserSearchRequest struct {
	Nickname string `json:"nickname"`
	Limit    int64  `json:"limit"`
}

type UserSearchResponse struct {
	UserBriefInfo UserBriefInfo
	Nickname      string `json:"nickname" bson:"nickname"`
}
