package service

import (
	"auth-api/internal/model"
	"auth-api/pkg/broker"
	"auth-api/pkg/logger"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationService struct {
	rabbitMQ *broker.RabbitMQ
}

func NewNotificationService(rabbitMQ *broker.RabbitMQ) *NotificationService {
	return &NotificationService{rabbitMQ: rabbitMQ}
}

func (s *NotificationService) SendNotification(ctx context.Context, notRMQ model.NotificationRabbitMQ, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("failed to marshal data: %s", err.Error())
		return err
	}

	channel := s.rabbitMQ.Channels[broker.EXCHANGE_VERIFY_CODE]
	if channel == nil {
		logger.Errorf("Channel for exchange %s is nil", broker.EXCHANGE_VERIFY_CODE)
		return fmt.Errorf("channel is nil")
	}

	switch notRMQ.Exchange {
	case broker.EXCHANGE_VERIFY_CODE:
		err = s.rabbitMQ.Channels[broker.EXCHANGE_VERIFY_CODE].Publish(
			notRMQ.Exchange,   // exchange name
			notRMQ.RoutingKey, // routing key
			true,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        bytes,
			},
		)
		if err != nil {
			logger.Errorf("failed to publish to verify code exchange: %s", err.Error())
			return err
		}
	default:
		err := fmt.Errorf("unknown exchange: %s", notRMQ.Exchange)
		logger.Errorf(err.Error())
		return err
	}

	logger.Infof("Notification sent to %s with routing key %s", notRMQ.Exchange, notRMQ.RoutingKey)
	return nil
}
