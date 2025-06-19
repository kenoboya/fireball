package model

type NotificationRabbitMQ struct {
	Exchange   string
	RoutingKey string
}

type NotificationChat struct {
	Chat        ChatBriefInfo `json:"chat"`
	Sender      UserBriefInfo `json:"sender"`
	RecipientID string        `json:"recipient_id"`
}

type NotificationMessage struct {
	Chat        ChatBriefInfo    `json:"chat"`
	Message     MessageBriefInfo `json:"message"`
	Sender      UserBriefInfo    `json:"sender"`
	RecipientID string           `json:"recipient_id"`
}
