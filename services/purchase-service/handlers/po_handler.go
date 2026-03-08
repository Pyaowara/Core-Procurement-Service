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
	PRID      uint   `json:"pr_id" binding:"required"`
	VendorID  uint   `json:"vendor_id" binding:"required"`
	ItemIDs   []uint `json:"item_ids"` // Optional: if empty, use all PR items
	CreditDay int    `json:"credit_day"`
	DueDate   string `json:"due_date" binding:"required"`
}

type CreatePOItem struct {
	ItemName     string  `json:"item_name" binding:"required"`
	Description  string  `json:"description"`
	Quantity     int     `json:"quantity" binding:"required,gt=0"`
	Unit         string  `json:"unit" binding:"required"`
	PricePerUnit float64 `json:"price_per_unit" binding:"required"`
	Discount     float64 `json:"discount"`
	DiscountUnit string  `json:"discount_unit"`
	TotalPrice   float64 `json:"total_price"`
	RequiredDate string  `json:"required_date"`
}

type UpdatePORequest struct {
	Status    string         `json:"status"`
	CreditDay int            `json:"credit_day"`
	DueDate   string         `json:"due_date"`
	POItems   []CreatePOItem `json:"po_items"`
}

type ReceiveGoodsRequest struct {
	ReceivedQty map[string]int `json:"received_qty" binding:"required"`
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

	// Create map of all PR items for validation
	prItemsMap := make(map[uint]*models.PRItem)
	for i := range pr.Items {
		prItemsMap[pr.Items[i].ID] = &pr.Items[i]
	}

	// Create map of selected item IDs for quick lookup
	selectedItemsMap := make(map[uint]bool)
	if len(req.ItemIDs) > 0 {
		// Validate that all requested item IDs belong to this PR
		for _, itemID := range req.ItemIDs {
			if _, exists := prItemsMap[itemID]; !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "PR item ID " + fmt.Sprint(itemID) + " does not belong to this PR"})
				return
			}
			selectedItemsMap[itemID] = true
		}
	} else {
		// If no items specified, select all PR items
		for _, item := range pr.Items {
			selectedItemsMap[item.ID] = true
		}
	}

	// Check if there are any valid items to create PO
	validItemCount := 0
	for _, item := range pr.Items {
		if selectedItemsMap[item.ID] {
			validItemCount++
		}
	}

	// If no items to create, return error immediately
	if validItemCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid items to create PO"})
		return
	}

	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	// Use transaction to create PO, PO items, and vendor snapshot together
	var po models.PurchaseOrder
	var poCreationErr error
	poCreationErr = config.DB.Transaction(func(tx *gorm.DB) error {
		// Create PO
		po = models.PurchaseOrder{
			PONumber:  "PO_" + time.Now().Format("20060102150405"),
			PRID:      req.PRID,
			VendorID:  req.VendorID,
			Status:    models.POStatusDraft,
			CreditDay: req.CreditDay,
			DueDate:   dueDate,
		}

		if err := tx.Create(&po).Error; err != nil {
			return err
		}

		// Create PO items from selected PR items
		itemsCreated := 0
		for _, prItem := range pr.Items {
			if !selectedItemsMap[prItem.ID] {
				continue // Skip items not in selection
			}
			// Calculate total price automatically
			totalPrice := CalculateTotalPrice(prItem.Quantity, prItem.PricePerUnit, prItem.Discount, prItem.DiscountUnit)
			poItem := models.POItem{
				POID:         po.ID,
				ItemName:     prItem.ItemName,
				Description:  prItem.Description,
				Quantity:     prItem.Quantity,
				Unit:         prItem.Unit,
				PricePerUnit: prItem.PricePerUnit,
				Discount:     prItem.Discount,
				DiscountUnit: prItem.DiscountUnit,
				TotalPrice:   totalPrice,
				RequiredDate: prItem.RequiredDate,
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

		// Create vendor snapshot for data consistency
		vendorSnapshotData, _ := json.Marshal(vendor)
		vendorSnapshot := models.VendorSnapshot{
			POID:          po.ID,
			VendorID:      vendor.ID,
			VendorName:    vendor.Name,
			VendorAddress: vendor.Address,
			VendorTaxID:   vendor.TaxID,
			SnapshotData:  vendorSnapshotData,
		}
		if err := tx.Create(&vendorSnapshot).Error; err != nil {
			return err
		}

		return nil
	})

	if poCreationErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create PO and items: " + poCreationErr.Error()})
		return
	}

	// Reload PO with items
	config.DB.Preload("Items").First(&po, po.ID)

	// Publish PO_CREATED event
	event := messaging.POCreatedEvent{
		POID:       po.ID,
		PONumber:   po.PONumber,
		PRID:       po.PRID,
		VendorID:   po.VendorID,
		VendorName: vendor.Name,
		DueDate:    dueDate.Format("2006-01-02"),
		Timestamp:  time.Now(),
	}

	for _, item := range po.Items {
		event.Items = append(event.Items, messaging.POItemPayload{
			ItemName:     item.ItemName,
			Description:  item.Description,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			PricePerUnit: item.PricePerUnit,
			Discount:     item.Discount,
			DiscountUnit: item.DiscountUnit,
			TotalPrice:   item.TotalPrice,
			RequiredDate: item.RequiredDate.Format("2006-01-02"),
		})
	}

	eventBytes, _ := json.Marshal(event)
	if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventPOCreated, eventBytes); err != nil {
		log.Printf("failed to publish PO_CREATED event: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "PO created successfully",
		"data":    po,
	})
}

// UpdatePOStatus updates the PO status
func UpdatePOStatus(c *gin.Context) {
	poID := c.Param("id")

	var req UpdatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var po models.PurchaseOrder
	if err := config.DB.First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	if req.Status != "" {
		validStatuses := map[string]bool{
			string(models.POStatusDraft):     true,
			string(models.POStatusSent):      true,
			string(models.POStatusCompleted): true,
		}
		if !validStatuses[req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		po.Status = models.POStatus(req.Status)
	}

	if req.CreditDay > 0 {
		po.CreditDay = req.CreditDay
	}

	if req.DueDate != "" {
		dueDate, _ := time.Parse("2006-01-02", req.DueDate)
		po.DueDate = dueDate
	}

	config.DB.Save(&po)

	c.JSON(http.StatusOK, gin.H{
		"message": "PO updated successfully",
		"data":    po,
	})
}

// GetPO retrieves a PO by ID
func GetPO(c *gin.Context) {
	poID := c.Param("id")

	var po models.PurchaseOrder
	if err := config.DB.Preload("Items").Preload("VendorSnapshot").First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	c.JSON(http.StatusOK, po)
}

// GetPOList retrieves all POs
func GetPOList(c *gin.Context) {
	status := c.Query("status")
	prID := c.Query("pr_id")

	var pos []models.PurchaseOrder
	query := config.DB

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

// ReceiveGoods records goods reception
func ReceiveGoods(c *gin.Context) {
	poID := c.Param("id")

	var req ReceiveGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var po models.PurchaseOrder
	if err := config.DB.Preload("Items").First(&po, poID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}

	// Create goods received record
	receivedData, _ := json.Marshal(req.ReceivedQty)
	goodsReceived := models.GoodsReceived{
		POID:         po.ID,
		ReceivedData: receivedData,
		ReceivedAt:   time.Now(),
	}
	config.DB.Create(&goodsReceived)

	// Publish GOODS_RECEIVED event
	event := messaging.GoodsReceivedEvent{
		POID:        po.ID,
		PONumber:    po.PONumber,
		ReceivedQty: req.ReceivedQty,
		Timestamp:   time.Now(),
	}

	eventBytes, _ := json.Marshal(event)
	if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventGoodsReceived, eventBytes); err != nil {
		log.Printf("failed to publish GOODS_RECEIVED event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Goods received successfully",
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
