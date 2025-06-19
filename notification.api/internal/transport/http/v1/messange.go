package v1

import (
	"context"
	"encoding/json"
	"notification-api/internal/model"
	"notification-api/pkg/broker"
	"notification-api/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (h *Handler) consumeSendMessage(ctx context.Context) {
	const consumerName = "SendMessage"

	ch := h.services.RabbitMQ.Channels[broker.EXCHANGE_MESSAGE]
	if ch == nil {
		logger.Errorf("[%s] Channel is not initialized", consumerName)
		return
	}

	msgs, err := ch.Consume(
		broker.QUEUE_MESSAGE_SEND,
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

			h.processSendMessage(ctx, msg, consumerName)
		}
	}
}

func (h *Handler) processSendMessage(ctx context.Context, msg amqp.Delivery, consumerName string) {
	logger.Debugf("[%s] Received message, size: %d bytes", consumerName, len(msg.Body))

	var nm model.NotificationMessage
	if err := json.Unmarshal(msg.Body, &nm); err != nil {
		logger.Errorf("[%s] Failed to unmarshal message: %v", consumerName, err)
		msg.Nack(false, false)
		return
	}

	nm.MessageAction = model.MESSAGE_ACTION_SEND

	if err := h.services.Messages.SaveNotificationMessage(ctx, nm); err != nil {
		logger.Errorf("[%s] Failed to save notification message: %v", consumerName, err)

		if h.shouldRequeue(err) {
			logger.Infof("[%s] Requeuing message for retry", consumerName)
			msg.Nack(false, true)
		} else {
			logger.Infof("[%s] Discarding message (permanent error)", consumerName)
			msg.Nack(false, false)
		}
		return
	}

	logger.Infof("[%s] Successfully processed message for chat_id: %d", consumerName, nm.Chat.ChatID)
	msg.Ack(false)
}

// TODO consumeMessageSendEncrypted
