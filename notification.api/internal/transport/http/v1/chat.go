package v1

import (
	"context"
	"encoding/json"
	"notification-api/internal/model"
	"notification-api/pkg/broker"
	"notification-api/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (h *Handler) consumeCreateChat(ctx context.Context) {
	const consumerName = "CreateChat"

	ch := h.services.RabbitMQ.Channels[broker.EXCHANGE_CHAT]
	if ch == nil {
		logger.Errorf("[%s] Channel is not initialized", consumerName)
		return
	}

	msgs, err := ch.Consume(
		broker.QUEUE_CHAT_CREATED,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Errorf("[%s] Failed to register consumer: %v", consumerName, err)
		return
	}

	logger.Infof("[%s] Consumer registered, waiting for messages...", consumerName)

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("[%s] Panic recovered: %v", consumerName, r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Infof("[%s] Context cancelled, stopping consumer", consumerName)
			return

		case msg, ok := <-msgs:
			if !ok {
				logger.Warnf("[%s] Message channel closed, stopping consumer", consumerName)
				return
			}

			h.processCreateChat(ctx, msg, consumerName)
		}
	}
}

func (h *Handler) processCreateChat(ctx context.Context, msg amqp.Delivery, consumerName string) {
	logger.Debugf("[%s] Received message, size: %d bytes", consumerName, len(msg.Body))

	var chat model.NotificationChat
	if err := json.Unmarshal(msg.Body, &chat); err != nil {
		logger.Errorf("[%s] Failed to unmarshal message: %v", consumerName, err)
		msg.Nack(false, false)
		return
	}

	chat.ChatAction = "private " + model.CHAT_ACTION_CREATE

	if err := h.services.Chats.SaveNotificationChat(ctx, chat); err != nil {
		logger.Errorf("[%s] Failed to save notification chat: %v", consumerName, err)

		if h.shouldRequeue(err) {
			logger.Infof("[%s] Requeuing message for retry", consumerName)
			msg.Nack(false, true)
		} else {
			logger.Infof("[%s] Discarding message (permanent error)", consumerName)
			msg.Nack(false, false)
		}
		return
	}

	logger.Infof("[%s] Successfully processed chat creation for chat_id: %d", consumerName, chat.Chat.ChatID)
	msg.Ack(false)
}

// TODO consumeMessageSendEncrypted
