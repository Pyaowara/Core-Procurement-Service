package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/messaging"
	"github.com/core-procurement/purchase-service/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePORequest struct {
	PRID      uint           `json:"pr_id" binding:"required"`
	VendorID  uint           `json:"vendor_id" binding:"required"`
	POItems   []CreatePOItem `json:"po_items" binding:"required,gt=0"` // Select PR Items and set custom values
	CreditDay int            `json:"credit_day"`
	DueDate   string         `json:"due_date" binding:"required"`
}

type CreatePOItem struct {
	PRItemID     uint     `json:"pr_item_id" binding:"required"` // Which PR Item to use
	SKU          string   `json:"sku"`                           // Required if PR Item SKU is empty
	ItemName     string   `json:"item_name"`                     // Required if PR Item ItemName is empty
	Description  string   `json:"description"`                   // Required if PR Item Description is empty
	Quantity     *int     `json:"quantity"`                      // Optional: override from PR Item
	PricePerUnit *float64 `json:"price_per_unit"`                // Optional: override from PR Item
	Discount     *float64 `json:"discount"`                      // Optional: override from PR Item
	DiscountUnit *string  `json:"discount_unit"`                 // Optional: override from PR Item
	RequiredDate *string  `json:"required_date"`                 // Optional: override from PR Item
}

type ReceiveGoodsRequest struct {
	// Empty - goods received based on PO item quantities automatically
}

// CalculateTotalPrice calculates total price based on quantity, price per unit, and discount
func CalculateTotalPrice(quantity int, pricePerUnit float64, discount float64, discountUnit string) float64 {
	subtotal := float64(quantity) * pricePerUnit
	if discountUnit == "%" {
		return subtotal * (1 - discount/100)
	}
	// Discount in BAHT
	return subtotal - discount
}

// ValidatePOItem validates that required fields are provided when PR Item values are empty
// Returns validation error if required fields are missing
func ValidatePOItem(item CreatePOItem, prItem *models.PRItem, index int) error {
	// Validate SKU - must be provided if PR Item SKU is empty
	sku := item.SKU
	if sku == "" {
		sku = prItem.SKU
	}
	if sku == "" {
		return fmt.Errorf("item %d: SKU cannot be empty (must provide in request if PR Item SKU is empty)", index+1)
	}

	// Validate ItemName - must be provided if PR Item ItemName is empty
	itemName := item.ItemName
	if itemName == "" {
		itemName = prItem.ItemName
	}
	if itemName == "" {
		return fmt.Errorf("item %d: item_name cannot be empty (must provide in request if PR Item ItemName is empty)", index+1)
	}

	// Validate Description - must be provided if PR Item Description is empty
	description := item.Description
	if description == "" {
		description = prItem.Description
	}
	if description == "" {
		return fmt.Errorf("item %d: description cannot be empty (must provide in request if PR Item Description is empty)", index+1)
	}

	// Validate discount for BAHT currency - must not exceed subtotal
	// Get final values (request override or PR Item default)
	quantity := prItem.Quantity
	if item.Quantity != nil {
		quantity = *item.Quantity
	}

	pricePerUnit := prItem.PricePerUnit
	if item.PricePerUnit != nil {
		pricePerUnit = *item.PricePerUnit
	}

	discountUnit := prItem.DiscountUnit
	if item.DiscountUnit != nil {
		discountUnit = *item.DiscountUnit
	}

	discount := prItem.Discount
	if item.Discount != nil {
		discount = *item.Discount
	}

	// If discount unit is BAHT, validate that discount doesn't exceed subtotal
	if discountUnit == "BAHT" {
		subtotal := float64(quantity) * pricePerUnit
		if discount > subtotal {
			return fmt.Errorf("item %d: BAHT discount (%.2f) cannot exceed item subtotal (%.2f)", index+1, discount, subtotal)
		}
	}

	return nil
}

// GeneratePO creates a Purchase Order from an approved PR with transaction support
// If there are no valid items to create, the PO will not be created
func GeneratePO(c *gin.Context) {
	var req CreatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify PR exists and is approved
	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, req.PRID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	if pr.Status != models.PRStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only approved PRs can generate PO"})
		return
	}

	// Verify vendor exists
	var vendor models.Vendor
	if err := config.DB.First(&vendor, req.VendorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
		return
	}

	// Validate that all requested PR Item IDs belong to this PR
	prItemsMap := make(map[uint]*models.PRItem)
	for i := range pr.Items {
		prItemsMap[pr.Items[i].ID] = &pr.Items[i]
	}

	for _, poItem := range req.POItems {
		if _, exists := prItemsMap[poItem.PRItemID]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PR item ID " + fmt.Sprint(poItem.PRItemID) + " does not belong to this PR"})
			return
		}
	}

	// Check if there are PO items to create
	if len(req.POItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one PO item is required"})
		return
	}

	// Validate all PO items before creating - check SKU, ItemName, Description requirements
	for i, poItem := range req.POItems {
		prItem := prItemsMap[poItem.PRItemID]
		if err := ValidatePOItem(poItem, prItem, i); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "PO creation failed - validation error on items",
			})
			return
		}
	}

	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	// Use transaction to create PO and PO items together
	var po models.PurchaseOrder
	var poCreationErr error
	poCreationErr = config.DB.Transaction(func(tx *gorm.DB) error {
		// Get the PR info for Purpose
		var pr models.PurchaseRequest
		if err := tx.First(&pr, req.PRID).Error; err != nil {
			return fmt.Errorf("failed to fetch PR: %v", err)
		}

		// Create PO
		po = models.PurchaseOrder{
			PONumber:  "PO_" + time.Now().Format("20060102150405"),
			PRID:      req.PRID,
			VendorID:  req.VendorID,
			Purpose:   pr.Purpose,
			Status:    models.POStatusSent,
			CreditDay: req.CreditDay,
			DueDate:   dueDate,
		}

		if err := tx.Create(&po).Error; err != nil {
			return err
		}

		// Create PO items from request with PR Item as base
		itemsCreated := 0
		for _, reqItem := range req.POItems {
			// Get PR Item details
			prItem := prItemsMap[reqItem.PRItemID]

			// Use request values or default to PR Item values
			sku := reqItem.SKU
			if sku == "" {
				sku = prItem.SKU
			}

			itemName := reqItem.ItemName
			if itemName == "" {
				itemName = prItem.ItemName
			}

			description := reqItem.Description
			if description == "" {
				description = prItem.Description
			}

			quantity := prItem.Quantity
			if reqItem.Quantity != nil {
				quantity = *reqItem.Quantity
			}

			pricePerUnit := prItem.PricePerUnit
			if reqItem.PricePerUnit != nil {
				pricePerUnit = *reqItem.PricePerUnit
			}

			discount := prItem.Discount
			if reqItem.Discount != nil {
				discount = *reqItem.Discount
			}

			discountUnit := prItem.DiscountUnit
			if reqItem.DiscountUnit != nil {
				discountUnit = *reqItem.DiscountUnit
			}

			requiredDate := prItem.RequiredDate
			if reqItem.RequiredDate != nil {
				parsedDate, err := time.Parse("2006-01-02", *reqItem.RequiredDate)
				if err == nil {
					requiredDate = parsedDate
				}
			}

			// Calculate total price automatically
			totalPrice := CalculateTotalPrice(quantity, pricePerUnit, discount, discountUnit)

			poItem := models.POItem{
				POID:         po.ID,
				SKU:          sku,
				ItemName:     itemName,
				Description:  description,
				Quantity:     quantity,
				PricePerUnit: pricePerUnit,
				Discount:     discount,
				DiscountUnit: discountUnit,
				TotalPrice:   totalPrice,
				RequiredDate: requiredDate,
			}
			if err := tx.Create(&poItem).Error; err != nil {
				return err
			}
			itemsCreated++
		}

		// If no items were created, rollback transaction
		if itemsCreated == 0 {
			return fmt.Errorf("no PO items created")
		}

		return nil
	})

	if poCreationErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create PO and items: " + poCreationErr.Error()})
		return
	}

	// Reload PO with items
	config.DB.Preload("Items").First(&po, po.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "PO created successfully",
		"data":    po,
	})
}

// GetPO retrieves a PO by ID with role-based filtering
func GetPO(c *gin.Context) {
	poID := c.Param("id")
	role, _ := c.Get("role")

	var po models.PurchaseOrder
	if err := config.DB.Preload("Items").First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	// Non-PurchaseOfficer and non-Admin roles can only view non-deleted POs
	if role.(string) != "PurchaseOfficer" && role.(string) != "Admin" {
		if po.IsDeleted {
			c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
			return
		}
	}

	c.JSON(http.StatusOK, po)
}

// GetPOList retrieves POs with role-based filtering
func GetPOList(c *gin.Context) {
	status := c.Query("status")
	prID := c.Query("pr_id")
	role, _ := c.Get("role")

	var pos []models.PurchaseOrder
	query := config.DB

	// Non-PurchaseOfficer and non-Admin roles can only view non-deleted POs
	if role.(string) != "PurchaseOfficer" && role.(string) != "Admin" {
		query = query.Where("is_deleted = ?", false)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if prID != "" {
		query = query.Where("pr_id = ?", prID)
	}

	if err := query.Preload("Items").Find(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve POs"})
		return
	}

	c.JSON(http.StatusOK, pos)
}

// ReceiveGoods records goods reception based on PO item quantities
// Only requires PO ID - updates all items as received with their full quantities
// Can only be used when PO status is SENT
func ReceiveGoods(c *gin.Context) {
	poID := c.Param("id")

	var po models.PurchaseOrder
	if err := config.DB.Preload("Items").First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	// Check if PO status is SENT
	if po.Status == models.POStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This PO has already received goods."})
		return
	}

	// Build received quantities from PO items
	receivedQty := make(map[string]int)
	for _, item := range po.Items {
		receivedQty[item.SKU] = item.Quantity
	}

	// Create goods received record
	receivedData, _ := json.Marshal(receivedQty)
	goodsReceived := models.GoodsReceived{
		POID:         po.ID,
		ReceivedData: receivedData,
		ReceivedAt:   time.Now(),
	}
	config.DB.Create(&goodsReceived)

	// Update PO status to COMPLETED
	po.Status = models.POStatusCompleted
	config.DB.Save(&po)

	// Publish GOODS_RECEIVED event with item details to inventory service
	event := messaging.GoodsReceivedEvent{
		POID:      po.ID,
		PONumber:  po.PONumber,
		Items:     []messaging.GoodsReceivedItem{},
		Timestamp: time.Now(),
	}

	// Add items from PO to event
	for _, item := range po.Items {
		event.Items = append(event.Items, messaging.GoodsReceivedItem{
			SKU:         item.SKU,
			ItemName:    item.ItemName,
			Description: item.Description,
			Quantity:    item.Quantity,
		})
	}

	eventBytes, _ := json.Marshal(event)
	if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventGoodsReceived, eventBytes); err != nil {
		log.Printf("failed to publish GOODS_RECEIVED event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Goods received successfully and PO marked as COMPLETED",
		"data":    goodsReceived,
	})
}

// DeletePO soft deletes a PO
func DeletePO(c *gin.Context) {
	poID := c.Param("id")

	var po models.PurchaseOrder
	if err := config.DB.First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	po.IsDeleted = true
	config.DB.Save(&po)

	c.JSON(http.StatusOK, gin.H{"message": "PO deleted successfully"})
}
