package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/core-procurement/inventory-service/config"
	"github.com/core-procurement/inventory-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
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

	var (
		conn *amqp.Connection
		err  error
	)
	for i := 1; i <= 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ not ready (attempt %d/10): %v — retrying in 5s...", i, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after 10 attempts: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	MQClient = &RabbitMQ{Conn: conn, Channel: ch}
	log.Println("RabbitMQ connection established (inventory-service)")
	return nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

// DeclareExchange creates a topic exchange if it doesn't exist
func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.Channel.ExchangeDeclare(name, "topic", true, false, false, false, nil)
}

// DeclareQueue creates a durable queue
func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(name, true, false, false, false, nil)
}

// BindQueue binds a queue to an exchange with a routing key
func (r *RabbitMQ) BindQueue(queueName, exchangeName, routingKey string) error {
	return r.Channel.QueueBind(queueName, routingKey, exchangeName, false, nil)
}

// PublishMessage sends a message to the exchange
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

// StartGoodsReceivedConsumer listens for goods.received events and updates inventory.
// On DB error it publishes inventory.update.failed back to purchase-service.
func StartGoodsReceivedConsumer() {
	q, err := MQClient.DeclareQueue("inventory-goods-received-queue")
	if err != nil {
		log.Printf("inventory consumer: failed to declare queue: %v", err)
		return
	}

	if err := MQClient.BindQueue(q.Name, ExchangeName, EventGoodsReceived); err != nil {
		log.Printf("inventory consumer: failed to bind queue: %v", err)
		return
	}

	msgs, err := MQClient.Channel.Consume(
		q.Name,
		"",
		false, // manual ack — we ack only on success
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("inventory consumer: failed to start consuming: %v", err)
		return
	}

	log.Println("inventory-service: listening for goods.received events...")
	for msg := range msgs {
		var event GoodsReceivedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("inventory consumer: bad message body: %v", err)
			msg.Nack(false, false) // dead-letter, don't requeue
			continue
		}

		if err := processGoodsReceived(event); err != nil {
			log.Printf("inventory consumer: failed to update inventory for PO %d: %v", event.POID, err)
			msg.Nack(false, false)
			publishFailureEvent(event, err.Error())
		} else {
			msg.Ack(false)
			log.Printf("inventory consumer: inventory updated for PO %d (%s)", event.POID, event.PONumber)
		}
	}
}

// processGoodsReceived increments stock for each item in the event inside a single DB transaction.
// If a SKU doesn't exist yet, a new inventory record is created (upsert behaviour).
func processGoodsReceived(event GoodsReceivedEvent) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range event.Items {
			var inv models.Inventory
			err := tx.Where("sku = ?", item.SKU).First(&inv).Error

			if err != nil {
				// SKU not found — create a new inventory record
				inv = models.Inventory{
					Sku:         item.SKU,
					Name:        item.ItemName,
					Description: item.Description,
					Quantity:    item.Quantity,
				}
				if createErr := tx.Create(&inv).Error; createErr != nil {
					return fmt.Errorf("failed to create inventory for SKU %s: %w", item.SKU, createErr)
				}
			} else {
				// SKU found — increment quantity
				if updateErr := tx.Model(&inv).Update("quantity", inv.Quantity+item.Quantity).Error; updateErr != nil {
					return fmt.Errorf("failed to update inventory for SKU %s: %w", item.SKU, updateErr)
				}
			}
		}
		return nil
	})
}

// publishFailureEvent tells purchase-service to mark the PO as FAILED
func publishFailureEvent(event GoodsReceivedEvent, reason string) {
	failEvent := InventoryUpdateFailedEvent{
		POID:      event.POID,
		PONumber:  event.PONumber,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	body, err := json.Marshal(failEvent)
	if err != nil {
		log.Printf("inventory consumer: failed to marshal failure event: %v", err)
		return
	}
	if err := MQClient.PublishMessage(ExchangeName, EventInventoryUpdateFailed, body); err != nil {
		log.Printf("inventory consumer: failed to publish failure event: %v", err)
	} else {
		log.Printf("inventory consumer: published inventory.update.failed for PO %d", event.POID)
	}
}
