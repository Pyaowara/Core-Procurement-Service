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
		// Try to unmarshal as ApprovalCompletedEvent first
		var completedEvent messaging.ApprovalCompletedEvent
		if err := json.Unmarshal(msg.Body, &completedEvent); err == nil && completedEvent.PRID != 0 {
			handleApprovalCompleted(completedEvent)
		} else {
			// Try to unmarshal as ApprovalRejectedEvent
			var rejectedEvent messaging.ApprovalRejectedEvent
			if err := json.Unmarshal(msg.Body, &rejectedEvent); err == nil && rejectedEvent.PRID != 0 {
				handleApprovalRejected(rejectedEvent)
			}
		}
	}
}

// SubscribeToInventoryEvents listens for inventory events (for future use)
func SubscribeToInventoryEvents() {
	// Declare queue
	q, err := messaging.MQClient.DeclareQueue("purchase-inventory-queue")
	if err != nil {
		log.Printf("failed to declare inventory queue: %v", err)
		return
	}

	// You can bind to additional inventory events here if needed
	// For now, just demonstrate the structure
	if err := messaging.MQClient.BindQueue(q.Name, messaging.ExchangeName, "inventory.*"); err != nil {
		log.Printf("failed to bind inventory queue: %v", err)
	}

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
		log.Printf("failed to consume inventory messages: %v", err)
		return
	}

	log.Println("Listening for inventory events...")
	for msg := range msgs {
		log.Printf("Received inventory event: %s", string(msg.Body))
	}
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
