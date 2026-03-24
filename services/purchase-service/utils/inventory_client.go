package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// InventoryItemResponse represents response from Inventory Service
type InventoryItemResponse struct {
	ID          uint   `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// GetAllInventories fetches all inventory items from Inventory Service
func GetAllInventories(authToken string) ([]InventoryItemResponse, error) {
	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryServiceURL == "" {
		// Default URL for Docker environment - use service name instead of localhost
		inventoryServiceURL = "http://inventory-service:6768"
	}

	endpoint := fmt.Sprintf("%s/inventory", inventoryServiceURL)
	log.Printf("Fetching inventory from: %s", endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return nil, err
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	// Use HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to call inventory service: %v", err)
		return nil, fmt.Errorf("failed to connect to inventory service: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read inventory response: %v", err)
		return nil, fmt.Errorf("failed to read inventory data: %v", err)
	}

	var inventories []InventoryItemResponse
	if err := json.Unmarshal(body, &inventories); err != nil {
		log.Printf("failed to parse inventory response: %v", err)
		return nil, fmt.Errorf("failed to parse inventory data: %v", err)
	}

	log.Printf("Fetched %d inventory items from inventory service", len(inventories))
	return inventories, nil
}

// GetInventoryItemBySKU fetches item details by SKU from Inventory Service
func GetInventoryItemBySKU(sku string, authToken string) (*InventoryItemResponse, error) {
	inventories, err := GetAllInventories(authToken)
	if err != nil {
		return nil, err
	}

	// Search for item by SKU
	skuLower := strings.ToLower(strings.TrimSpace(sku))
	for i := range inventories {
		if strings.ToLower(strings.TrimSpace(inventories[i].SKU)) == skuLower {
			log.Printf("Found item with SKU '%s': %s", sku, inventories[i].Name)
			return &inventories[i], nil
		}
	}

	log.Printf("Item with SKU '%s' not found in inventory", sku)
	return nil, fmt.Errorf("SKU not found in inventory")
}
