package service

import (
	"chat-api/internal/model"
	"chat-api/pkg/broker"
	"chat-api/pkg/logger"
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

	switch notRMQ.Exchange {
	case broker.EXCHANGE_CHAT:
		err = s.rabbitMQ.Channels[broker.EXCHANGE_CHAT].Publish(
			notRMQ.Exchange,
			notRMQ.RoutingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        bytes,
			},
		)
		if err != nil {
			logger.Errorf("failed to publish to chat exchange: %s", err.Error())
			return err
		}

	case broker.EXCHANGE_MESSAGE:
		err = s.rabbitMQ.Channels[broker.EXCHANGE_MESSAGE].Publish(
			notRMQ.Exchange,
			notRMQ.RoutingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        bytes,
			},
		)
		if err != nil {
			logger.Errorf("failed to publish to message exchange: %s", err.Error())
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
