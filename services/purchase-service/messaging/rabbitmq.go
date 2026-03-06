package messaging

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

var MQClient *RabbitMQ

func ConnectRabbitMQ() error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Printf("failed to connect to RabbitMQ: %v", err)
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open channel: %v", err)
		return err
	}

	MQClient = &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}

	log.Println("RabbitMQ connection established")
	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return err
	}
	return r.Conn.Close()
}

// DeclareExchange creates an exchange if it doesn't exist
func (r *RabbitMQ) DeclareExchange(exchangeName string) error {
	return r.Channel.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

// DeclareQueue creates a queue if it doesn't exist
func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
}

// BindQueue binds a queue to an exchange
func (r *RabbitMQ) BindQueue(queueName, exchangeName, routingKey string) error {
	return r.Channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}

// PublishMessage publishes a message to an exchange
func (r *RabbitMQ) PublishMessage(exchangeName, routingKey string, body []byte) error {
	return r.Channel.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
