package services

import (
	"encoding/json"
	"log"

	"github.com/core-procurement/approval-service/messaging"
)

// SubscribeToPREvents listens for PR events from purchase-service
func SubscribeToPREvents() {
	// Declare queue
	q, err := messaging.MQClient.DeclareQueue("approval-pr-queue")
	if err != nil {
		log.Printf("failed to declare pr queue: %v", err)
		return
	}

	// Bind queue to exchange for pr.ready.for.approval event
	if err := messaging.MQClient.BindQueue(q.Name, messaging.ExchangeName, messaging.EventPRReadyForApproval); err != nil {
		log.Printf("failed to bind pr.ready.for.approval queue: %v", err)
		return
	}

	// Consume messages
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
		log.Printf("failed to consume pr messages: %v", err)
		return
	}

	log.Println("Listening for PR events...")
	for msg := range msgs {
		var event messaging.PRReadyForApprovalEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("failed to unmarshal PR event: %v", err)
			continue
		}

		handlePRReadyForApproval(event)
	}
}

// handlePRReadyForApproval automatically creates an approval workflow when a PR is submitted
func handlePRReadyForApproval(event messaging.PRReadyForApprovalEvent) {
	log.Printf("Handling PR ready for approval: PR ID=%d, PR Number=%s, WorkflowID=%s", event.PRID, event.PRNumber, event.WorkflowID)

	// Check if approval workflow already exists
	existingInstance, err := GetApprovalInstance("PR", event.PRID)
	if err == nil && existingInstance != nil {
		log.Printf("Approval workflow already exists for PR %d, skipping creation", event.PRID)
		return
	}

	// Create new approval workflow with role-based approvals
	instance, err := CreateApprovalInstance("PR", event.PRID, event.RequesterID, event.WorkflowID)
	if err != nil {
		log.Printf("failed to create approval instance for PR %d: %v", event.PRID, err)
		return
	}

	log.Printf("Created approval workflow (ID=%d, WorkflowID=%s) for PR %d with 4 steps", instance.ID, instance.WorkflowID, event.PRID)
}
