package services

import (
	"encoding/json"
	"log"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/messaging"
	"github.com/core-procurement/purchase-service/models"
)

// SubscribeToApprovalEvents listens for approval events from approval-service
func SubscribeToApprovalEvents() {
	// Declare queue
	q, err := messaging.MQClient.DeclareQueue("purchase-approval-queue")
	if err != nil {
		log.Printf("failed to declare approval queue: %v", err)
		return
	}

	// Bind queue to exchange for approval completed event
	if err := messaging.MQClient.BindQueue(q.Name, messaging.ExchangeName, messaging.EventApprovalCompleted); err != nil {
		log.Printf("failed to bind approval completed queue: %v", err)
	}

	// Bind queue to exchange for approval rejected event
	if err := messaging.MQClient.BindQueue(q.Name, messaging.ExchangeName, messaging.EventApprovalRejected); err != nil {
		log.Printf("failed to bind approval rejected queue: %v", err)
	}

	// Consume messages
	msgs, err := messaging.MQClient.Channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to consume approval messages: %v", err)
		return
	}

	log.Println("Listening for approval events...")
	for msg := range msgs {
		switch msg.RoutingKey {
		case messaging.EventApprovalCompleted:
			var completedEvent messaging.ApprovalCompletedEvent
			if err := json.Unmarshal(msg.Body, &completedEvent); err != nil {
				log.Printf("failed to unmarshal approval completed event: %v", err)
				continue
			}
			if completedEvent.PRID == 0 {
				log.Printf("invalid approval completed event: missing pr_id")
				continue
			}
			handleApprovalCompleted(completedEvent)

		case messaging.EventApprovalRejected:
			var rejectedEvent messaging.ApprovalRejectedEvent
			if err := json.Unmarshal(msg.Body, &rejectedEvent); err != nil {
				log.Printf("failed to unmarshal approval rejected event: %v", err)
				continue
			}
			if rejectedEvent.PRID == 0 {
				log.Printf("invalid approval rejected event: missing pr_id")
				continue
			}
			handleApprovalRejected(rejectedEvent)

		default:
			log.Printf("ignoring unknown approval routing key: %s", msg.RoutingKey)
		}
	}
}

// SubscribeToInventoryEvents listens for inventory events from inventory-service.
// Handles: inventory.update.failed — marks the affected PO as FAILED.
func SubscribeToInventoryEvents() {
	q, err := messaging.MQClient.DeclareQueue("inventory-goods-received-queue")
	if err != nil {
		log.Printf("failed to declare inventory queue: %v", err)
		return
	}

	if err := messaging.MQClient.BindQueue(q.Name, messaging.ExchangeName, messaging.EventInventoryUpdateFailed); err != nil {
		log.Printf("failed to bind inventory.update.failed queue: %v", err)
		return
	}

	msgs, err := messaging.MQClient.Channel.Consume(
		q.Name,
		"",
		true, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to consume inventory messages: %v", err)
		return
	}

	log.Println("Listening for inventory events...")
	for msg := range msgs {
		var event messaging.InventoryUpdateFailedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("failed to unmarshal inventory event: %v", err)
			continue
		}
		if event.POID != 0 {
			handleInventoryUpdateFailed(event)
		}
	}
}

// handleInventoryUpdateFailed marks the PO as FAILED when inventory-service reports an error.
func handleInventoryUpdateFailed(event messaging.InventoryUpdateFailedEvent) {
	log.Printf("Handling inventory update failure for PO ID: %d, reason: %s", event.POID, event.Reason)

	var po models.PurchaseOrder
	if err := config.DB.First(&po, event.POID).Error; err != nil {
		log.Printf("failed to find PO %d: %v", event.POID, err)
		return
	}

	po.Status = models.POStatusFailed
	if err := config.DB.Save(&po).Error; err != nil {
		log.Printf("failed to update PO %d status to FAILED: %v", event.POID, err)
		return
	}

	log.Printf("PO %d status updated to FAILED due to inventory update failure", event.POID)
}

// handleApprovalCompleted updates PR status to APPROVED
func handleApprovalCompleted(event messaging.ApprovalCompletedEvent) {
	log.Printf("Handling approval completed for PR ID: %d", event.PRID)

	var pr models.PurchaseRequest
	if err := config.DB.First(&pr, event.PRID).Error; err != nil {
		log.Printf("failed to find PR: %v", err)
		return
	}

	pr.Status = models.PRStatusApproved
	if err := config.DB.Save(&pr).Error; err != nil {
		log.Printf("failed to update PR status: %v", err)
		return
	}

	log.Printf("PR %d status updated to APPROVED", event.PRID)
}

// handleApprovalRejected updates PR status to REJECTED
func handleApprovalRejected(event messaging.ApprovalRejectedEvent) {
	log.Printf("Handling approval rejected for PR ID: %d, Reason: %s", event.PRID, event.Reason)

	var pr models.PurchaseRequest
	if err := config.DB.First(&pr, event.PRID).Error; err != nil {
		log.Printf("failed to find PR: %v", err)
		return
	}

	pr.Status = models.PRStatusRejected
	if err := config.DB.Save(&pr).Error; err != nil {
		log.Printf("failed to update PR status: %v", err)
		return
	}

	log.Printf("PR %d status updated to REJECTED", event.PRID)
}
