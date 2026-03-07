package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/messaging"
	"github.com/core-procurement/purchase-service/models"
	"github.com/core-procurement/purchase-service/utils"
	"github.com/gin-gonic/gin"
)

type CreatePRRequest struct {
	PRNumber   string                `json:"pr_number" binding:"required"`
	Department string                `json:"department" binding:"required"`
	Items      []CreatePRItemRequest `json:"items" binding:"required"`
}

type CreatePRItemRequest struct {
	ItemName     string  `json:"item_name" binding:"required"`
	Description  string  `json:"description"`
	Quantity     int     `json:"quantity" binding:"required,gt=0"`
	Unit         string  `json:"unit" binding:"required"`
	PricePerUnit float64 `json:"price_per_unit"`
	Discount     float64 `json:"discount"`
	DiscountUnit string  `json:"discount_unit"`
	RequiredDate string  `json:"required_date"`
}

type UpdatePRRequest struct {
	Department string                `json:"department"`
	Items      []CreatePRItemRequest `json:"items"`
}

// CreatePR creates a new Purchase Request in DRAFT status
func CreatePR(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	pr := models.PurchaseRequest{
		PRNumber:    req.PRNumber,
		RequesterID: userID.(uint),
		Department:  req.Department,
		Status:      models.PRStatusDraft,
	}

	if err := config.DB.Create(&pr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create PR"})
		return
	}

	// Create PR items with auto-calculated total price
	for _, item := range req.Items {
		requiredDate, _ := time.Parse("2006-01-02", item.RequiredDate)
		// Calculate total price automatically
		totalPrice := CalculateTotalPrice(item.Quantity, item.PricePerUnit, item.Discount, item.DiscountUnit)
		prItem := models.PRItem{
			PRID:         pr.ID,
			ItemName:     item.ItemName,
			Description:  item.Description,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			PricePerUnit: item.PricePerUnit,
			Discount:     item.Discount,
			DiscountUnit: item.DiscountUnit,
			TotalPrice:   totalPrice,
			RequiredDate: requiredDate,
		}
		config.DB.Create(&prItem)
	}

	// Reload PR with items
	config.DB.Preload("Items").First(&pr, pr.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "PR created successfully",
		"data":    pr,
	})
}

// UpdatePR updates PR (only possible in DRAFT status)
func UpdatePR(c *gin.Context) {
	prID := c.Param("id")

	var req UpdatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pr models.PurchaseRequest
	if err := config.DB.First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Only allow updating DRAFT PRs
	if pr.Status != models.PRStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only DRAFT PRs can be updated"})
		return
	}

	// Update department
	if req.Department != "" {
		pr.Department = req.Department
	}

	config.DB.Save(&pr)

	// Delete old items and create new ones
	if len(req.Items) > 0 {
		config.DB.Where("pr_id = ?", pr.ID).Delete(&models.PRItem{})

		for _, item := range req.Items {
			requiredDate, _ := time.Parse("2006-01-02", item.RequiredDate)
			// Calculate total price automatically
			totalPrice := CalculateTotalPrice(item.Quantity, item.PricePerUnit, item.Discount, item.DiscountUnit)
			prItem := models.PRItem{
				PRID:         pr.ID,
				ItemName:     item.ItemName,
				Description:  item.Description,
				Quantity:     item.Quantity,
				Unit:         item.Unit,
				PricePerUnit: item.PricePerUnit,
				Discount:     item.Discount,
				DiscountUnit: item.DiscountUnit,
				TotalPrice:   totalPrice,
				RequiredDate: requiredDate,
			}
			config.DB.Create(&prItem)
		}
	}

	// Reload PR with items
	config.DB.Preload("Items").First(&pr, pr.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "PR updated successfully",
		"data":    pr,
	})
}

// GetPR retrieves a PR by ID
func GetPR(c *gin.Context) {
	prID := c.Param("id")

	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	c.JSON(http.StatusOK, pr)
}

// GetPRList retrieves all PRs for the current user
func GetPRList(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")

	var prs []models.PurchaseRequest
	query := config.DB.Where("requester_id = ?", userID.(uint))

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Preload("Items").Find(&prs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve PRs"})
		return
	}

	c.JSON(http.StatusOK, prs)
}

// SubmitPR submits PR for approval with full validation and inventory checks
// Workflow:
// 1. Validate Data: Check that items exist and have required fields
// 2. Check Inventory Availability: Query Inventory Service for current stock levels
// 3. Take Inventory Snapshot: Capture current stock state in PR Items
// 4. Change Status: Update PR status from DRAFT to PENDING
// 5. Trigger Approval: Generate workflow ID and publish PR_READY_FOR_APPROVAL event
func SubmitPR(c *gin.Context) {
	prID := c.Param("id")

	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Step 1: Validate Data
	if pr.Status != models.PRStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only DRAFT PRs can be submitted"})
		return
	}

	// Validate that PR has at least 1 item
	if len(pr.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PR must have at least 1 item"})
		return
	}

	// Validate that all items have required_date
	validationErrors := []string{}
	for i, item := range pr.Items {
		if item.RequiredDate.IsZero() {
			validationErrors = append(validationErrors, fmt.Sprintf("Item %d (%s): required_date must be specified", i+1, item.ItemName))
		}
		if item.Quantity <= 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("Item %d (%s): quantity must be greater than 0", i+1, item.ItemName))
		}
	}

	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": validationErrors,
		})
		return
	}

	// Step 2: Check Inventory Availability
	itemNames := make([]string, len(pr.Items))
	for i, item := range pr.Items {
		itemNames[i] = item.ItemName
	}

	log.Printf("Checking inventory for PR %d with items: %v", pr.ID, itemNames)
	stockCheckResults := utils.CheckInventoryStock(itemNames)

	// Collect inventory check details for response
	inventoryCheckDetails := make(map[string]interface{})
	hasErrors := false

	for itemName, checkResult := range stockCheckResults {
		if checkResult.Error != "" {
			log.Printf("Stock check warning for item %s: %s", itemName, checkResult.Error)
			inventoryCheckDetails[itemName] = gin.H{
				"available_qty": checkResult.AvailableQty,
				"warning":       checkResult.Error,
			}
			hasErrors = true // Warning but continue processing
		} else {
			inventoryCheckDetails[itemName] = gin.H{
				"available_qty": checkResult.AvailableQty,
				"checked_at":    checkResult.CheckedAt,
			}
		}
	}

	// Step 3: Take Inventory Snapshot - Update PR Items with current stock info
	now := time.Now()
	for i := range pr.Items {
		itemName := pr.Items[i].ItemName
		if result, exists := stockCheckResults[itemName]; exists {
			pr.Items[i].CurrentStockAtSubmit = result.AvailableQty
			pr.Items[i].StockCheckAt = now
		}
		// Save updated item
		if err := config.DB.Model(&pr.Items[i]).Updates(pr.Items[i]).Error; err != nil {
			log.Printf("failed to update PR item with stock info: %v", err)
		}
	}

	// Create inventory snapshot with all PR items
	snapshotData, _ := json.Marshal(pr.Items)
	snapshot := models.InventorySnapshot{
		PRID:         pr.ID,
		SnapshotData: snapshotData,
	}
	if err := config.DB.Create(&snapshot).Error; err != nil {
		log.Printf("failed to create inventory snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create snapshot"})
		return
	}

	log.Printf("Created inventory snapshot for PR %d", pr.ID)

	// Step 4: Change Status to PENDING
	pr.Status = models.PRStatusPending
	// Generate workflow ID with timestamp
	pr.WorkflowID = "WF_" + strconv.FormatUint(uint64(pr.ID), 10) + "_" + time.Now().Format("20060102150405")

	if err := config.DB.Save(&pr).Error; err != nil {
		log.Printf("failed to update PR status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PR status"})
		return
	}

	log.Printf("Updated PR %d status to PENDING with WorkflowID: %s", pr.ID, pr.WorkflowID)

	// Step 5: Trigger Approval by publishing event
	event := messaging.PRReadyForApprovalEvent{
		PRID:        pr.ID,
		PRNumber:    pr.PRNumber,
		RequesterID: pr.RequesterID,
		Department:  pr.Department,
		WorkflowID:  pr.WorkflowID,
		Timestamp:   time.Now(),
	}

	// Convert items with snapshot data
	for _, item := range pr.Items {
		event.Items = append(event.Items, messaging.PRItemPayload{
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
	if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventPRReadyForApproval, eventBytes); err != nil {
		log.Printf("failed to publish PR_READY_FOR_APPROVAL event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish approval event"})
		return
	}

	log.Printf("Published PR_READY_FOR_APPROVAL event for PR %d", pr.ID)

	// Return success response with inventory check details
	c.JSON(http.StatusOK, gin.H{
		"message":                  "PR submitted successfully",
		"data":                     pr,
		"inventory_check_summary":  inventoryCheckDetails,
		"has_inventory_warnings":   hasErrors,
		"snapshot_created":         true,
		"workflow_id":              pr.WorkflowID,
		"approval_event_published": true,
	})
}

// DeletePR soft deletes a PR
func DeletePR(c *gin.Context) {
	prID := c.Param("id")

	var pr models.PurchaseRequest
	if err := config.DB.First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	pr.IsDeleted = true
	config.DB.Save(&pr)

	c.JSON(http.StatusOK, gin.H{"message": "PR deleted successfully"})
}

// GetPRSnapshot retrieves the inventory snapshot for a PR (used for audit trail)
// Shows the state of items at the time the PR was submitted
func GetPRSnapshot(c *gin.Context) {
	prID := c.Param("id")

	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Get snapshot from database
	var snapshot models.InventorySnapshot
	if err := config.DB.Where("pr_id = ?", prID).First(&snapshot).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found for this PR"})
		return
	}

	// Parse snapshot data
	var snapshotItems []models.PRItem
	if err := json.Unmarshal(snapshot.SnapshotData, &snapshotItems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse snapshot"})
		return
	}

	// Compare snapshot with current PR items to identify changes
	type ItemComparison struct {
		SnapshotData  models.PRItem `json:"snapshot_data"`
		CurrentData   models.PRItem `json:"current_data"`
		HasChanged    bool          `json:"has_changed"`
		ChangedFields []string      `json:"changed_fields"`
	}

	comparisons := make([]ItemComparison, 0)
	for i, snapshotItem := range snapshotItems {
		comparison := ItemComparison{
			SnapshotData:  snapshotItem,
			HasChanged:    false,
			ChangedFields: []string{},
		}

		if i < len(pr.Items) {
			comparison.CurrentData = pr.Items[i]

			// Check for changes
			if snapshotItem.ItemName != pr.Items[i].ItemName {
				comparison.HasChanged = true
				comparison.ChangedFields = append(comparison.ChangedFields, "item_name")
			}
			if snapshotItem.Quantity != pr.Items[i].Quantity {
				comparison.HasChanged = true
				comparison.ChangedFields = append(comparison.ChangedFields, "quantity")
			}
			if snapshotItem.PricePerUnit != pr.Items[i].PricePerUnit {
				comparison.HasChanged = true
				comparison.ChangedFields = append(comparison.ChangedFields, "price_per_unit")
			}
			if snapshotItem.RequiredDate != pr.Items[i].RequiredDate {
				comparison.HasChanged = true
				comparison.ChangedFields = append(comparison.ChangedFields, "required_date")
			}
		}

		comparisons = append(comparisons, comparison)
	}

	c.JSON(http.StatusOK, gin.H{
		"pr_id":            pr.ID,
		"pr_number":        pr.PRNumber,
		"status":           pr.Status,
		"snapshot_created": snapshot.CreatedAt,
		"items_comparison": comparisons,
		"summary": gin.H{
			"total_items":   len(snapshotItems),
			"items_changed": len(comparisons),
		},
	})
}
