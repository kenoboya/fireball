package v1

import (
	"context"
	"encoding/json"
	"notification-api/internal/model"
	"notification-api/pkg/broker"
	"notification-api/pkg/logger"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (h *Handler) consumeVerifyCodeEmail(ctx context.Context) {
	const consumerName = "VerifyCodeEmail"

	ch := h.services.RabbitMQ.Channels[broker.EXCHANGE_VERIFY_CODE]
	if ch == nil {
		logger.Errorf("[%s] Channel is not initialized", consumerName)
		return
	}

	msgs, err := ch.Consume(
		broker.QUEUE_VERIFY_CODE_SEND_TO_EMAIL,
		"",    // consumer tag
		false, // auto-ack - лучше делать manual ack для надежности
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
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

			h.processVerifyCodeEmailMessage(ctx, msg, consumerName)
		}
	}
}

func (h *Handler) processVerifyCodeEmailMessage(ctx context.Context, msg amqp.Delivery, consumerName string) {
	logger.Debugf("[%s] Received message, size: %d bytes", consumerName, len(msg.Body))

	var vc model.VerifyCodeInput
	if err := json.Unmarshal(msg.Body, &vc); err != nil {
		logger.Errorf("[%s] Failed to unmarshal message: %v", consumerName, err)
		msg.Nack(false, false)
		return
	}

	if err := h.services.Emails.SendVerifyCodeToEmail(ctx, vc); err != nil {
		if h.shouldRequeue(err) {
			logger.Infof("[%s] Requeuing message for retry", consumerName)
			msg.Nack(false, true)
		} else {
			logger.Infof("[%s] Discarding message (permanent error)", consumerName)
			msg.Nack(false, false)
		}
		return
	}

	logger.Infof("[%s] Successfully sent verification email to: %s", consumerName, vc.Recipient)
	msg.Ack(false)
}

func (h *Handler) shouldRequeue(err error) bool {
	errorStr := err.Error()
	temporaryErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
	}

	for _, tempErr := range temporaryErrors {
		if strings.Contains(strings.ToLower(errorStr), tempErr) {
			return true
		}
	}

	return false
}

func (h *Handler) consumeVerifyCodePhone(ctx context.Context) {
	const consumerName = "VerifyCodePhone"

	ch := h.services.RabbitMQ.Channels[broker.EXCHANGE_VERIFY_CODE]
	if ch == nil {
		logger.Errorf("[%s] Channel is not initialized", consumerName)
		return
	}

	msgs, err := ch.Consume(
		broker.QUEUE_VERIFY_CODE_SEND_TO_PHONE,
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

			h.processVerifyCodePhoneMessage(ctx, msg, consumerName)
		}
	}
}

func (h *Handler) processVerifyCodePhoneMessage(ctx context.Context, msg amqp.Delivery, consumerName string) {
	logger.Debugf("[%s] Received message, size: %d bytes", consumerName, len(msg.Body))

	var vc model.VerifyCodeInput
	if err := json.Unmarshal(msg.Body, &vc); err != nil {
		logger.Errorf("[%s] Failed to unmarshal message: %v", consumerName, err)
		msg.Nack(false, false)
		return
	}

	logger.Infof("[%s] Processing verification code for phone: %s", consumerName, vc.Recipient)

	if err := h.services.Phones.SendVerifyCodeToPhone(ctx, vc); err != nil {
		logger.Errorf("[%s] Failed to send SMS to %s: %v", consumerName, vc.Recipient, err)

		if h.shouldRequeue(err) {
			logger.Infof("[%s] Requeuing message for retry", consumerName)
			msg.Nack(false, true)
		} else {
			logger.Infof("[%s] Discarding message (permanent error)", consumerName)
			msg.Nack(false, false)
		}
		return
	}

	logger.Infof("[%s] Successfully sent SMS to: %s", consumerName, vc.Recipient)
	msg.Ack(false)
}
