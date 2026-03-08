package messaging

import (
	"encoding/json"
	"time"
)

// Event types
const (
	ExchangeName = "procurement"

	// Events published by approval-service
	EventApprovalCompleted = "approval.completed"
	EventApprovalRejected  = "approval.rejected"

	// Events subscribed from other services
	EventPRReadyForApproval = "pr.ready.for.approval"
)

// ApprovalCompletedEvent is published when approval workflow is fully approved
type ApprovalCompletedEvent struct {
	PRID       uint      `json:"pr_id"`
	WorkflowID string    `json:"workflow_id"`
	Status     string    `json:"status"`
	ApprovedAt time.Time `json:"approved_at"`
}

// ApprovalRejectedEvent is published when approval workflow is rejected
type ApprovalRejectedEvent struct {
	PRID       uint      `json:"pr_id"`
	WorkflowID string    `json:"workflow_id"`
	Reason     string    `json:"reason"`
	RejectedAt time.Time `json:"rejected_at"`
}

// PRReadyForApprovalEvent is received when purchase-service submits a PR
type PRReadyForApprovalEvent struct {
	PRID        uint      `json:"pr_id"`
	PRNumber    string    `json:"pr_number"`
	RequesterID uint      `json:"requester_id"`
	Department  string    `json:"department"`
	WorkflowID  string    `json:"workflow_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// Helper function to marshal events to JSON
func MarshalEvent(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}
