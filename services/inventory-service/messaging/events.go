package messaging

import "time"

// Event routing keys
const (
	ExchangeName = "procurement"

	// Events consumed by inventory-service
	EventGoodsReceived = "goods.received"

	// Events published by inventory-service
	EventInventoryUpdateFailed = "inventory.update.failed"
)

// GoodsReceivedEvent is published by purchase-service when goods are received
type GoodsReceivedEvent struct {
	POID      uint                `json:"po_id"`
	PONumber  string              `json:"po_number"`
	Items     []GoodsReceivedItem `json:"items"`
	Timestamp time.Time           `json:"timestamp"`
}

type GoodsReceivedItem struct {
	SKU         string `json:"sku"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// InventoryUpdateFailedEvent is published by inventory-service when a stock update fails
type InventoryUpdateFailedEvent struct {
	POID      uint      `json:"po_id"`
	PONumber  string    `json:"po_number"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}
