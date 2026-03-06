package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// InventoryItemResponse represents response from Inventory Service
type InventoryItemResponse struct {
	ID          uint   `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// StockCheckResult holds the current stock information
type StockCheckResult struct {
	ItemName     string
	AvailableQty int
	CheckedAt    string
	Error        string
}

// CheckInventoryStock queries the Inventory Service for current stock levels
// Returns map of item_name -> available_qty
func CheckInventoryStock(itemNames []string) map[string]StockCheckResult {
	result := make(map[string]StockCheckResult)

	// Get inventory service URL from env or use default
	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryServiceURL == "" {
		inventoryServiceURL = "http://localhost:6768"
	}

	// Get all inventories
	endpoint := fmt.Sprintf("%s/inventory", inventoryServiceURL)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("failed to call inventory service: %v", err)
		for _, item := range itemNames {
			result[item] = StockCheckResult{
				ItemName: item,
				Error:    fmt.Sprintf("failed to connect to inventory service: %v", err),
			}
		}
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read inventory response: %v", err)
		for _, item := range itemNames {
			result[item] = StockCheckResult{
				ItemName: item,
				Error:    fmt.Sprintf("failed to read inventory data: %v", err),
			}
		}
		return result
	}

	var inventories []InventoryItemResponse
	if err := json.Unmarshal(body, &inventories); err != nil {
		log.Printf("failed to parse inventory response: %v", err)
		for _, item := range itemNames {
			result[item] = StockCheckResult{
				ItemName: item,
				Error:    fmt.Sprintf("failed to parse inventory data: %v", err),
			}
		}
		return result
	}

	// Create map of inventory items for quick lookup
	inventoryMap := make(map[string]InventoryItemResponse)
	for _, inv := range inventories {
		inventoryMap[strings.ToLower(inv.Name)] = inv
	}

	// Check each requested item
	for _, itemName := range itemNames {
		itemKey := strings.ToLower(itemName)
		if inv, exists := inventoryMap[itemKey]; exists {
			result[itemName] = StockCheckResult{
				ItemName:     itemName,
				AvailableQty: inv.Quantity,
				CheckedAt:    "now",
			}
		} else {
			result[itemName] = StockCheckResult{
				ItemName:     itemName,
				AvailableQty: 0,
				Error:        "item not found in inventory",
			}
		}
	}

	return result
}

// CheckSingleItemStock checks stock for a single item
func CheckSingleItemStock(itemName string) StockCheckResult {
	results := CheckInventoryStock([]string{itemName})
	if result, exists := results[itemName]; exists {
		return result
	}
	return StockCheckResult{
		ItemName: itemName,
		Error:    "unknown error",
	}
}
