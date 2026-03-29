package messaging

import (
	"encoding/json"
	"time"
)

// Event types
const (
	ExchangeName = "procurement"

	// Events published by purchase-service
	EventPRReadyForApproval = "pr.ready.for.approval"
	EventPOCreated          = "po.created"
	EventGoodsReceived      = "goods.received"

	// Events subscribed from other services
	EventApprovalCompleted     = "approval.completed"
	EventApprovalRejected      = "approval.rejected"
	EventInventoryUpdateFailed = "inventory.update.failed"
)

// PRReadyForApprovalEvent is published when PR is submitted and ready for approval
type PRReadyForApprovalEvent struct {
	PRID        uint            `json:"pr_id"`
	PRNumber    string          `json:"pr_number"`
	RequesterID uint            `json:"requester_id"`
	Department  string          `json:"department"`
	Purpose     string          `json:"purpose"`
	Items       []PRItemPayload `json:"items"`
	WorkflowID  string          `json:"workflow_id"`
	Timestamp   time.Time       `json:"timestamp"`
}

type PRItemPayload struct {
	SKU          string  `json:"sku"`
	ItemName     string  `json:"item_name"`
	Description  string  `json:"description"`
	Quantity     int     `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
	Discount     float64 `json:"discount"`
	DiscountUnit string  `json:"discount_unit"`
	TotalPrice   float64 `json:"total_price"`
	RequiredDate string  `json:"required_date"`
}

// ApprovalCompletedEvent is received when approval service completes approval
type ApprovalCompletedEvent struct {
	PRID       uint      `json:"pr_id"`
	WorkflowID string    `json:"workflow_id"`
	Status     string    `json:"status"`
	ApprovedAt time.Time `json:"approved_at"`
}

// ApprovalRejectedEvent is received when approval service rejects PR
type ApprovalRejectedEvent struct {
	PRID       uint      `json:"pr_id"`
	WorkflowID string    `json:"workflow_id"`
	Reason     string    `json:"reason"`
	RejectedAt time.Time `json:"rejected_at"`
}


// GoodsReceivedEvent is published when goods are received - includes items for inventory update
type GoodsReceivedEvent struct {
	POID      uint                `json:"po_id"`
	PONumber  string              `json:"po_number"`
	Items     []GoodsReceivedItem `json:"items"` // List of items received
	Timestamp time.Time           `json:"timestamp"`
}

type GoodsReceivedItem struct {
	SKU         string `json:"sku"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// InventoryUpdateFailedEvent is published by inventory-service when it fails to update inventory
type InventoryUpdateFailedEvent struct {
	POID      uint      `json:"po_id"`
	PONumber  string    `json:"po_number"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// Helper function to marshal events to JSON
func MarshalEvent(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}

// Helper function to unmarshal events from JSON
func UnmarshalEvent(data []byte, event interface{}) error {
	return json.Unmarshal(data, event)
}
