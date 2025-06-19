package model

const (
	EMAIL = "email"
	SMS   = "sms"
)

type NotificationRabbitMQ struct {
	Exchange   string
	RoutingKey string
}

type NotificationChat struct {
	Chat        ChatBriefInfo `json:"chat"`
	Sender      UserBriefInfo `json:"sender"`
	RecipientID string        `json:"recipient_id"`
	ChatAction  string
}

type NotificationMessage struct {
	Chat          ChatBriefInfo    `json:"chat"`
	Message       MessageBriefInfo `json:"message"`
	Sender        UserBriefInfo    `json:"sender"`
	RecipientID   string           `json:"recipient_id"`
	MessageAction string
}

type VerifyCodeInput struct {
	Recipient string `json:"recipient"`
	Code      string `json:"code"`
}

type NotificationResponse struct {
	NotificationMessages []NotificationMessage `json:"new_messages"`
	NotificationChat     []NotificationChat    `json:"new_chats"`
	MutedChat            []ChatBriefInfo       `json:"muted_chat"`
}
