package broker

import (
	"auth-api/pkg/logger"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EXCHANGE_VERIFY_CODE = "verify_code_exchange"

	QUEUE_VERIFY_CODE_SEND_TO_PHONE = "verify_code_send_to_phone_queue"
	QUEUE_VERIFY_CODE_SEND_TO_EMAIL = "verify_code_send_to_email_queue"

	ROUTING_KEY_VERIFY_CODE_EMAIL = "verify_code.email"
	ROUTING_KEY_VERIFY_CODE_PHONE = "verify_code.phone"
)

type RabbitMQConfig struct {
	Host     string `envconfig:"HOST"`
	Port     int    `envconfig:"PORT"`
	User     string `envconfig:"USER"`
	Password string `envconfig:"PASSWORD"`
}

type RabbitMQ struct {
	conn     *amqp.Connection
	Channels map[string]*amqp.Channel
}

func NewRabbitMQ(config RabbitMQConfig) (*RabbitMQ, error) {
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%d/", config.User, config.Password, config.Host, config.Port)

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &RabbitMQ{
		conn:     conn,
		Channels: make(map[string]*amqp.Channel),
	}, nil
}

func (r *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	return r.conn.Channel()
}

func (r *RabbitMQ) CloseConnection() error {
	return r.conn.Close()
}

func (r *RabbitMQ) InitializationOfChannels() error {
	var err error
	if err = r.initializationOfVerifyCodeChannel(); err != nil {
		logger.Errorf("Failed to initialize verify code channel: %s", err.Error())
		return err
	}

	return nil
}

func (r *RabbitMQ) initializationOfVerifyCodeChannel() error {
	var ch1 *amqp.Channel
	var err error

	ch1, err = r.conn.Channel()
	if err != nil {
		logger.Errorf("Failed to open verify code channel: %s", err.Error())
		return err
	}

	err = ch1.ExchangeDeclare(
		EXCHANGE_VERIFY_CODE, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		logger.Errorf("Failed to create verify_code_exchange: %s", err.Error())
		return err
	}

	queueBindings := map[string]string{
		QUEUE_VERIFY_CODE_SEND_TO_EMAIL: ROUTING_KEY_VERIFY_CODE_EMAIL,
		QUEUE_VERIFY_CODE_SEND_TO_PHONE: ROUTING_KEY_VERIFY_CODE_PHONE,
	}

	for queueName, routingKey := range queueBindings {
		// Declare queue
		q, err := ch1.QueueDeclare(
			queueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			logger.Errorf("Failed to declare queue %s: %s", queueName, err.Error())
			return err
		}

		// Bind queue with specific routing key
		err = ch1.QueueBind(
			q.Name,               // queue name
			routingKey,           // routing key (verify_code.email or verify_code.phone)
			EXCHANGE_VERIFY_CODE, // exchange
			false,                // no-wait
			nil,                  // arguments
		)
		if err != nil {
			logger.Errorf("Failed to bind queue %s to exchange with routing key %s: %s", queueName, routingKey, err.Error())
			return err
		}

		logger.Infof("Queue %s bound to exchange %s with routing key %s", queueName, EXCHANGE_VERIFY_CODE, routingKey)
	}

	r.Channels[EXCHANGE_VERIFY_CODE] = ch1
	return nil
}
