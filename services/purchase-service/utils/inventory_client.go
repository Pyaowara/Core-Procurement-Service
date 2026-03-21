package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/models"
	"gorm.io/gorm"
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
		inventoryServiceURL = "http://localhost:6768"
	}

	endpoint := fmt.Sprintf("%s/inventory", inventoryServiceURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return nil, err
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
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

// FetchInventoryItemAndCreateSnapshot fetches all items from inventory service and creates a snapshot with all items
// Returns the requested item data, status message, and error
func FetchInventoryItemAndCreateSnapshot(sku string, authToken string) (*InventoryItemResponse, string, error) {
	// Try to fetch all inventory items from inventory service
	allInventories, err := GetAllInventories(authToken)
	if err != nil {
		// Inventory service failed - try to use snapshot as fallback
		log.Printf("Inventory service failed for SKU '%s': %v. Attempting to use snapshot.", sku, err)

		// Try to find the most recent snapshot
		snapshot, snapshotErr := GetLatestSnapshot()
		if snapshotErr != nil {
			// No snapshot available - fail
			errMsg := "SKU not found in inventory and no cached snapshot available"
			log.Printf("SKU '%s' failed - no inventory data and no snapshot: %v", sku, snapshotErr)
			return nil, "", fmt.Errorf("%s", errMsg)
		}

		// Use snapshot data - search for requested SKU in cached inventory list
		var inventoriesList []InventoryItemResponse
		if err := json.Unmarshal(snapshot.SnapshotData, &inventoriesList); err != nil {
			log.Printf("Failed to parse snapshot data for SKU '%s': %v", sku, err)
			return nil, "", fmt.Errorf("failed to parse cached snapshot")
		}

		// Find the requested SKU in cached list
		skuLower := strings.ToLower(strings.TrimSpace(sku))
		for i := range inventoriesList {
			if strings.ToLower(strings.TrimSpace(inventoriesList[i].SKU)) == skuLower {
				log.Printf("Using snapshot for SKU '%s': %s with quantity %d (snapshot created at %v)", sku, inventoriesList[i].Name, inventoriesList[i].Quantity, snapshot.CreatedAt)
				status := "fetched from snapshot (inventory service unavailable)"
				return &inventoriesList[i], status, nil
			}
		}

		log.Printf("SKU '%s' not found in cached snapshot inventory list", sku)
		return nil, "", fmt.Errorf("SKU not found in inventory snapshot")
	}

	// Find the requested item in the fetched list
	var requestedItem *InventoryItemResponse
	skuLower := strings.ToLower(strings.TrimSpace(sku))
	for i := range allInventories {
		if strings.ToLower(strings.TrimSpace(allInventories[i].SKU)) == skuLower {
			requestedItem = &allInventories[i]
			break
		}
	}

	if requestedItem == nil {
		log.Printf("SKU '%s' not found in inventory", sku)
		return nil, "", fmt.Errorf("SKU not found in inventory")
	}

	// Successfully fetched from inventory service - create new snapshot with ALL inventory items as list
	allInventoriesJSON, _ := json.Marshal(allInventories)

	newSnapshot := models.InventorySnapshot{
		SnapshotData: allInventoriesJSON,
	}
	if err := config.DB.Create(&newSnapshot).Error; err != nil {
		log.Printf("Warning: failed to create snapshot for SKU '%s': %v", sku, err)
		// Don't fail - still use the inventory item
	} else {
		log.Printf("New snapshot created for SKU '%s' - stored %d total inventory items (requested item quantity: %d)", sku, len(allInventories), requestedItem.Quantity)
	}

	log.Printf("Successfully fetched item for SKU '%s' from inventory: %s", sku, requestedItem.Name)
	status := "fetched from inventory"
	return requestedItem, status, nil
}

// GetLatestSnapshot retrieves the most recent inventory snapshot
// Contains all inventory items as a list - searches within it for requested SKU
// Used as fallback when inventory service is unavailable
func GetLatestSnapshot() (*models.InventorySnapshot, error) {
	var snapshot models.InventorySnapshot
	result := config.DB.
		Order("created_at DESC").
		First(&snapshot)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no snapshot available")
		}
		return nil, fmt.Errorf("failed to retrieve snapshot: %v", result.Error)
	}

	return &snapshot, nil
}
