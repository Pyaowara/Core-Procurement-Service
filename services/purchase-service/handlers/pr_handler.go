package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/messaging"
	"github.com/core-procurement/purchase-service/models"
	"github.com/core-procurement/purchase-service/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePRRequest struct {
	PRNumber   string                `json:"pr_number"`
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

// GeneratePRNumber generates a unique PR number based on PR ID and timestamp
// Format: PR-YYYYMMDD-{PRID:06d}
func GeneratePRNumber(prID uint) string {
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("PR-%s-%06d", timestamp, prID)
}

// CheckAndCreatePRItems checks inventory for each item and creates PR items only for insufficient stock
// Requires authToken for Inventory Service authentication
// Returns:
// - items to create with adjusted quantities
// - inventory check details for response
func CheckAndCreatePRItems(pr *models.PurchaseRequest, items []CreatePRItemRequest, authToken string) ([]models.PRItem, map[string]interface{}, error) {
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.ItemName
	}

	log.Printf("Checking inventory for PR with items: %v", itemNames)
	stockCheckResults := utils.CheckInventoryStock(itemNames, authToken)

	inventoryCheckDetails := make(map[string]interface{})
	var prItemsToCreate []models.PRItem

	for _, item := range items {
		checkResult := stockCheckResults[item.ItemName]
		availableQty := checkResult.AvailableQty

		// Calculate quantity needed for PR
		var prQuantity int
		var status string
		log.Printf("Inventory check for item '%s': requested %d, available %d", item.ItemName, item.Quantity, availableQty)
		if checkResult.Error != "" {
			// Item not found in inventory, create PR for full quantity
			prQuantity = item.Quantity
			status = "not found in inventory, creating PR for full quantity"
			log.Printf("Item %s not found in inventory. Creating PR for %d units", item.ItemName, prQuantity)
		} else if availableQty >= item.Quantity {
			// Enough stock available, don't create PR item
			prQuantity = 0
			status = fmt.Sprintf("sufficient stock (%d units available)", availableQty)
			log.Printf("Item %s has sufficient stock (%d available >= %d needed)", item.ItemName, availableQty, item.Quantity)
		} else {
			// Insufficient stock, create PR for the shortage
			prQuantity = item.Quantity - availableQty
			status = fmt.Sprintf("insufficient stock (%d available), creating PR for shortage of %d units", availableQty, prQuantity)
			log.Printf("Item %s has insufficient stock (%d available < %d needed). Creating PR for %d units", item.ItemName, availableQty, item.Quantity, prQuantity)
		}

		// Add to inventory check details
		inventoryCheckDetails[item.ItemName] = gin.H{
			"requested_qty":  item.Quantity,
			"available_qty":  availableQty,
			"pr_qty_created": prQuantity,
			"status":         status,
		}

		// Only create PR item if quantity > 0
		if prQuantity > 0 {
			requiredDate, _ := time.Parse("2006-01-02", item.RequiredDate)
			// Calculate total price automatically based on PR quantity (not requested quantity)
			totalPrice := CalculateTotalPrice(prQuantity, item.PricePerUnit, item.Discount, item.DiscountUnit)
			prItem := models.PRItem{
				PRID:         pr.ID,
				ItemName:     item.ItemName,
				Description:  item.Description,
				Quantity:     prQuantity,
				Unit:         item.Unit,
				PricePerUnit: item.PricePerUnit,
				Discount:     item.Discount,
				DiscountUnit: item.DiscountUnit,
				TotalPrice:   totalPrice,
				RequiredDate: requiredDate,
			}
			prItemsToCreate = append(prItemsToCreate, prItem)
		}
	}

	return prItemsToCreate, inventoryCheckDetails, nil
}

// CreatePR creates a new Purchase Request in DRAFT status with transaction
// Now includes inventory checking and only creates PR items for insufficient stock
// If no items are created, the PR will not be created (transaction rollback)
func CreatePR(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	// Get JWT token from Authorization header for inventory service calls
	authToken := c.GetHeader("Authorization")
	authToken = strings.TrimPrefix(authToken, "Bearer ")

	// Check inventory and determine which items need PR BEFORE creating PR
	tempPR := models.PurchaseRequest{
		RequesterID: userID.(uint),
		Department:  req.Department,
		Status:      models.PRStatusDraft,
	}
	prItemsToCreate, inventoryCheckDetails, err := CheckAndCreatePRItems(&tempPR, req.Items, authToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check inventory"})
		return
	}

	// Validate that there are items to create
	if len(prItemsToCreate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":                     "no items to create PR - all items have sufficient stock",
			"inventory_check_summary":   inventoryCheckDetails,
			"total_items_requested":     len(req.Items),
			"items_with_sufficient_qty": len(req.Items),
		})
		return
	}

	// Use transaction to ensure PR and PR items are created together
	var pr models.PurchaseRequest
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Create PR with provided PR number or empty (will be auto-generated)
		pr = models.PurchaseRequest{
			PRNumber:    req.PRNumber,
			RequesterID: userID.(uint),
			Department:  req.Department,
			Status:      models.PRStatusDraft,
		}

		if err := tx.Create(&pr).Error; err != nil {
			return err
		}

		// If PR number was not provided, generate one based on PR ID
		if req.PRNumber == "" {
			pr.PRNumber = GeneratePRNumber(pr.ID)
			if err := tx.Save(&pr).Error; err != nil {
				return err
			}
			log.Printf("Generated PR number: %s for PR ID: %d", pr.PRNumber, pr.ID)
		}

		// Create PR items with correct PR ID
		for _, prItem := range prItemsToCreate {
			prItem.PRID = pr.ID
			if err := tx.Create(&prItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create PR and items"})
		return
	}

	// Reload PR with items
	config.DB.Preload("Items").First(&pr, pr.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":                   "PR created successfully",
		"data":                      pr,
		"inventory_check_summary":   inventoryCheckDetails,
		"pr_items_created_count":    len(prItemsToCreate),
		"total_items_requested":     len(req.Items),
		"items_with_sufficient_qty": len(req.Items) - len(prItemsToCreate),
	})
}

// UpdatePR updates PR (only possible in DRAFT status) with transaction support
// Now includes inventory checking and only updates PR items for insufficient stock
// If no items are created, the update will be rolled back
func UpdatePR(c *gin.Context) {
	prID := c.Param("id")

	var req UpdatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get JWT token from Authorization header for inventory service calls
	authToken := c.GetHeader("Authorization")
	authToken = strings.TrimPrefix(authToken, "Bearer ")

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

	// If items are being updated, validate before transaction
	if len(req.Items) > 0 {
		// Check inventory and determine which items need PR BEFORE using transaction
		tempPR := models.PurchaseRequest{ID: pr.ID}
		prItemsToCreate, inventoryCheckDetails, err := CheckAndCreatePRItems(&tempPR, req.Items, authToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check inventory"})
			return
		}

		// Validate that there are items to create
		if len(prItemsToCreate) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                     "no items to update PR - all items have sufficient stock",
				"inventory_check_summary":   inventoryCheckDetails,
				"total_items_requested":     len(req.Items),
				"items_with_sufficient_qty": len(req.Items),
			})
			return
		}

		// Use transaction to ensure department update and items are created together
		var updatedPR models.PurchaseRequest
		err = config.DB.Transaction(func(tx *gorm.DB) error {
			// Update department
			if req.Department != "" {
				pr.Department = req.Department
				if err := tx.Save(&pr).Error; err != nil {
					return err
				}
			}

			// Delete old items
			if err := tx.Where("pr_id = ?", pr.ID).Delete(&models.PRItem{}).Error; err != nil {
				return err
			}

			// Create new PR items
			for _, prItem := range prItemsToCreate {
				prItem.PRID = pr.ID
				if err := tx.Create(&prItem).Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PR and items"})
			return
		}

		// Reload PR with items
		config.DB.Preload("Items").First(&updatedPR, pr.ID)

		c.JSON(http.StatusOK, gin.H{
			"message":                   "PR updated successfully",
			"data":                      updatedPR,
			"inventory_check_summary":   inventoryCheckDetails,
			"pr_items_created_count":    len(prItemsToCreate),
			"total_items_requested":     len(req.Items),
			"items_with_sufficient_qty": len(req.Items) - len(prItemsToCreate),
		})
	} else {
		// Update department only
		if req.Department != "" {
			pr.Department = req.Department
			config.DB.Save(&pr)
		}

		// Reload PR with items
		config.DB.Preload("Items").First(&pr, pr.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "PR updated successfully",
			"data":    pr,
		})
	}
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

// SubmitPR submits PR for approval
// Workflow:
// 1. Validate Data: Check that items exist and have required fields
// 2. Create Inventory Snapshot: Capture current state of PR items for audit
// 3. Change Status: Update PR status from DRAFT to PENDING
// 4. Trigger Approval: Generate workflow ID and publish PR_READY_FOR_APPROVAL event
// Note: Inventory checking is now done in CreatePR/UpdatePR, not here
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

	// Step 2: Create Inventory Snapshot - for audit trail
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

	// Step 3: Change Status to PENDING
	pr.Status = models.PRStatusPending
	// Generate workflow ID with timestamp
	pr.WorkflowID = "WF_" + strconv.FormatUint(uint64(pr.ID), 10) + "_" + time.Now().Format("20060102150405")

	if err := config.DB.Save(&pr).Error; err != nil {
		log.Printf("failed to update PR status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PR status"})
		return
	}

	log.Printf("Updated PR %d status to PENDING with WorkflowID: %s", pr.ID, pr.WorkflowID)

	// Step 4: Trigger Approval by publishing event
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

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":                  "PR submitted successfully",
		"data":                     pr,
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
