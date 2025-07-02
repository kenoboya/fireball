package broker

import (
	"fmt"
	"notification-api/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EXCHANGE_CHAT        = "chat_exchange"
	EXCHANGE_MESSAGE     = "message_exchange"
	EXCHANGE_VERIFY_CODE = "verify_code_exchange"

	QUEUE_CHAT_CREATED     = "chat_created"
	QUEUE_CHAT_DELETED     = "chat_deleted"
	QUEUE_CHAT_ADDED_USER  = "chat_added_user"
	QUEUE_CHAT_LEFT_USER   = "chat_left_user"
	QUEUE_CHAT_RENAME      = "chat_rename"
	QUEUE_CHAT_KICKED_USER = "chat_kicked_user"

	QUEUE_MESSAGE_SEND           = "message_send"
	QUEUE_MESSAGE_SEND_ENCRYPTED = "message_send_encrypted"

	QUEUE_VERIFY_CODE_SEND_TO_PHONE = "verify_code_send_to_phone_queue"
	QUEUE_VERIFY_CODE_SEND_TO_EMAIL = "verify_code_send_to_email_queue"

	ROUTING_KEY_VERIFY_CODE_EMAIL = "verify_code.email"
	ROUTING_KEY_VERIFY_CODE_PHONE = "verify_code.phone"

	ROUTING_KEY_CHAT_CREATED     = "chat.created"
	ROUTING_KEY_CHAT_DELETED     = "chat.deleted"
	ROUTING_KEY_CHAT_ADDED_USER  = "chat.user.added"
	ROUTING_KEY_CHAT_LEFT_USER   = "chat.user.left"
	ROUTING_KEY_CHAT_RENAME      = "chat.renamed"
	ROUTING_KEY_CHAT_KICKED_USER = "chat.user.kicked"

	ROUTING_KEY_MESSAGE_SEND           = "message.send"
	ROUTING_KEY_MESSAGE_SEND_ENCRYPTED = "message.send.encrypted"
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
	if err = r.initializationOfChatChannel(); err != nil {
		logger.Errorf("Failed to initialize chat channel: %s", err.Error())
		return err
	}

	if err = r.initializationOfMessageChannel(); err != nil {
		logger.Errorf("Failed to initialize message channel: %s", err.Error())
		return err
	}

	if err = r.initializationOfVerifyCodeChannel(); err != nil {
		logger.Errorf("Failed to initialize verify code channel: %s", err.Error())
		return err
	}
	return nil
}

func (r *RabbitMQ) initializationOfChatChannel() error {
	var ch1 *amqp.Channel
	var err error

	ch1, err = r.conn.Channel()
	if err != nil {
		logger.Errorf("Failed to open chat channel: %s", err.Error())
		return err
	}

	err = ch1.ExchangeDeclare(
		EXCHANGE_CHAT, // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		logger.Errorf("Failed to create %s: %s", EXCHANGE_CHAT, err.Error())
		return err
	}

	// Declare and bind queues with specific routing keys
	queueBindings := map[string]string{
		QUEUE_CHAT_CREATED:     ROUTING_KEY_CHAT_CREATED,
		QUEUE_CHAT_DELETED:     ROUTING_KEY_CHAT_DELETED,
		QUEUE_CHAT_ADDED_USER:  ROUTING_KEY_CHAT_ADDED_USER,
		QUEUE_CHAT_LEFT_USER:   ROUTING_KEY_CHAT_LEFT_USER,
		QUEUE_CHAT_RENAME:      ROUTING_KEY_CHAT_RENAME,
		QUEUE_CHAT_KICKED_USER: ROUTING_KEY_CHAT_KICKED_USER,
	}

	for queueName, routingKey := range queueBindings {
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

		err = ch1.QueueBind(
			q.Name,        // queue name
			routingKey,    // routing key
			EXCHANGE_CHAT, // exchange
			false,         // no-wait
			nil,           // arguments
		)
		if err != nil {
			logger.Errorf("Failed to bind queue %s to exchange with routing key %s: %s", queueName, routingKey, err.Error())
			return err
		}
	}
	r.Channels[EXCHANGE_CHAT] = ch1
	return nil
}

func (r *RabbitMQ) initializationOfMessageChannel() error {
	var ch2 *amqp.Channel
	var err error

	ch2, err = r.conn.Channel()
	if err != nil {
		logger.Errorf("Failed to open message channel: %s", err.Error())
		return err
	}

	err = ch2.ExchangeDeclare(
		EXCHANGE_MESSAGE, // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		logger.Errorf("Failed to create %s: %s", EXCHANGE_MESSAGE, err.Error())
		return err
	}

	// Declare and bind queues with specific routing keys
	queueBindings := map[string]string{
		QUEUE_MESSAGE_SEND:           ROUTING_KEY_MESSAGE_SEND,
		QUEUE_MESSAGE_SEND_ENCRYPTED: ROUTING_KEY_MESSAGE_SEND_ENCRYPTED,
	}

	for queueName, routingKey := range queueBindings {
		q, err := ch2.QueueDeclare(
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

		err = ch2.QueueBind(
			q.Name,           // queue name
			routingKey,       // routing key
			EXCHANGE_MESSAGE, // exchange
			false,            // no-wait
			nil,              // arguments
		)
		if err != nil {
			logger.Errorf("Failed to bind queue %s to exchange with routing key %s: %s", queueName, routingKey, err.Error())
			return err
		}
	}

	r.Channels[EXCHANGE_MESSAGE] = ch2
	return nil
}

func (r *RabbitMQ) initializationOfVerifyCodeChannel() error {
	var ch3 *amqp.Channel
	var err error

	ch3, err = r.conn.Channel()
	if err != nil {
		logger.Errorf("Failed to open verify code channel: %s", err.Error())
		return err
	}

	err = ch3.ExchangeDeclare(
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
		q, err := ch3.QueueDeclare(
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
		err = ch3.QueueBind(
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

	r.Channels[EXCHANGE_VERIFY_CODE] = ch3
	return nil
}
