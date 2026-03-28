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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePRRequest struct {
	PRNumber   string                `json:"pr_number"`
	Department string                `json:"department" binding:"required"`
	Purpose    string                `json:"purpose" binding:"required"`
	Items      []CreatePRItemRequest `json:"items" binding:"required"`
}

type CreatePRItemRequest struct {
	SKU          string  `json:"sku"` // Stock Keeping Unit - optional
	ItemName     string  `json:"item_name" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required,gt=0"`
	PricePerUnit float64 `json:"price_per_unit" binding:"required,gt=0"`
	Discount     float64 `json:"discount" binding:"min=0"`
	DiscountUnit string  `json:"discount_unit"`                    // Discount unit (e.g., percentage, %) - required only if discount > 0
	RequiredDate string  `json:"required_date" binding:"required"` // Required delivery date
}

type UpdatePRRequest struct {
	Department string                `json:"department" binding:"required"`
	Purpose    string                `json:"purpose" binding:"required"`
	Items      []CreatePRItemRequest `json:"items" binding:"required"`
}

// GeneratePRNumber generates a unique PR number based on PR ID and timestamp
// Format: PR-YYYYMMDD-{PRID:06d}
func GeneratePRNumber(prID uint) string {
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("PR-%s-%06d", timestamp, prID)
}

// ValidatePRItemRequest validates all fields in a PR item request for empty/whitespace values
func ValidatePRItemRequest(item CreatePRItemRequest, index int) error {

	if strings.TrimSpace(item.ItemName) == "" {
		return fmt.Errorf("item %d: item_name cannot be empty or whitespace", index+1)
	}

	if strings.TrimSpace(item.Description) == "" {
		return fmt.Errorf("item %d: description cannot be empty or whitespace", index+1)
	}

	if item.Quantity <= 0 {
		return fmt.Errorf("item %d: quantity must be greater than 0", index+1)
	}

	if item.PricePerUnit <= 0 {
		return fmt.Errorf("item %d: price_per_unit must be greater than 0", index+1)
	}

	if item.Discount < 0 {
		return fmt.Errorf("item %d: discount cannot be negative", index+1)
	}

	// If discount is provided (> 0), DiscountUnit must not be empty
	if item.Discount > 0 && strings.TrimSpace(item.DiscountUnit) == "" {
		return fmt.Errorf("item %d: discount_unit cannot be empty when discount is provided", index+1)
	}

	// If discount is 0, DiscountUnit is optional (can be empty)

	// Validate DiscountUnit - only allow "%" or "BAHT" when provided
	if strings.TrimSpace(item.DiscountUnit) != "" {
		discountUnit := strings.TrimSpace(item.DiscountUnit)
		if discountUnit != "%" && discountUnit != "BAHT" {
			return fmt.Errorf("item %d: discount_unit must be either '%s' or 'BAHT' (got: %s)", index+1, "%", discountUnit)
		}

		// If discount unit is BAHT, validate that discount doesn't exceed subtotal
		switch discountUnit {
		case "BAHT":
			subtotal := float64(item.Quantity) * item.PricePerUnit
			if item.Discount > subtotal {
				return fmt.Errorf("item %d: BAHT discount (%.2f) cannot exceed item subtotal (%.2f)", index+1, item.Discount, subtotal)
			}
		case "%":
			if item.Discount > 100 {
				return fmt.Errorf("item %d: percentage discount cannot exceed 100%% (got: %.2f%%)", index+1, item.Discount)
			}
		}
	}

	// Validate RequiredDate is not empty
	if strings.TrimSpace(item.RequiredDate) == "" {
		return fmt.Errorf("item %d: required_date cannot be empty", index+1)
	}

	// Validate RequiredDate format is valid
	_, err := time.Parse("2006-01-02", item.RequiredDate)
	if err != nil {
		return fmt.Errorf("item %d: required_date must be in format YYYY-MM-DD (got: %s)", index+1, item.RequiredDate)
	}

	return nil
}

func CheckAndCreatePRItems(pr *models.PurchaseRequest, items []CreatePRItemRequest) ([]models.PRItem, map[string]interface{}, error) {
	itemCheckDetails := make(map[string]interface{})
	var prItemsToCreate []models.PRItem
	var invalidItems []string

	for i, item := range items {
		// Validate all fields in the item (no SKU check required)
		if validationErr := ValidatePRItemRequest(item, i); validationErr != nil {
			log.Printf("Item validation failed for item %d: %v", i+1, validationErr)
			itemKey := fmt.Sprintf("item_%d", i)
			itemCheckDetails[itemKey] = gin.H{
				"sku":              item.SKU,
				"item_name":        item.ItemName,
				"description":      item.Description,
				"quantity":         item.Quantity,
				"status":           "failed",
				"validation_error": validationErr.Error(),
			}
			invalidItems = append(invalidItems, fmt.Sprintf("item %d: %s", i+1, validationErr.Error()))
			continue
		}

		// Validation passed - log and record
		log.Printf("Item validation passed for item %d", i+1)
		itemKey := fmt.Sprintf("item_%d", i)
		itemCheckDetails[itemKey] = gin.H{
			"sku":         item.SKU,
			"item_name":   item.ItemName,
			"description": item.Description,
			"quantity":    item.Quantity,
			"status":      "valid",
		}

		// Create PR item - SKU is optional
		requiredDate, _ := time.Parse("2006-01-02", item.RequiredDate)
		totalPrice := CalculateTotalPrice(item.Quantity, item.PricePerUnit, item.Discount, item.DiscountUnit)
		prItem := models.PRItem{
			PRID:         pr.ID,
			SKU:          item.SKU, // Can be empty
			ItemName:     item.ItemName,
			Description:  item.Description,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			Discount:     item.Discount,
			DiscountUnit: item.DiscountUnit,
			TotalPrice:   totalPrice,
			RequiredDate: requiredDate,
		}
		prItemsToCreate = append(prItemsToCreate, prItem)
	}

	// If any validations failed, fail the entire request
	if len(invalidItems) > 0 {
		errorMsg := fmt.Sprintf("Item validation failed: %s", strings.Join(invalidItems, "; "))
		log.Printf("%s", errorMsg)
		return nil, itemCheckDetails, fmt.Errorf("%s", errorMsg)
	}

	if len(prItemsToCreate) == 0 {
		return nil, itemCheckDetails, fmt.Errorf("no valid items found")
	}

	return prItemsToCreate, itemCheckDetails, nil
}

// CreatePR creates a new Purchase Request in DRAFT status with transaction
func CreatePR(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Role-based access control: only employee, manager, and admin can create PRs
	userRole, exists := c.Get("role")
	roleStr, ok := userRole.(string)
	if !exists || !ok || !strings.EqualFold(roleStr, "Employee") && !strings.EqualFold(roleStr, "Manager") && !strings.EqualFold(roleStr, "Admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only employees, managers, and admins can create Purchase Requests"})
		return
	}

	userID, _ := c.Get("user_id")

	// Validate items without inventory checks
	tempPR := models.PurchaseRequest{
		RequesterID: userID.(uint),
		Department:  req.Department,
		Purpose:     req.Purpose,
		Status:      models.PRStatusDraft,
	}
	prItemsToCreate, itemCheckDetails, err := CheckAndCreatePRItems(&tempPR, req.Items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":                   err.Error(),
			"message":                 "PR creation failed - validation error on items",
			"item_validation_summary": itemCheckDetails,
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
			Purpose:     req.Purpose,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create PR and items: " + err.Error()})
		return
	}

	// Reload PR with items
	config.DB.Preload("Items").First(&pr, pr.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":                 "PR created successfully",
		"data":                    pr,
		"item_validation_summary": itemCheckDetails,
		"pr_items_created_count":  len(prItemsToCreate),
	})
}

// UpdatePR updates PR (only possible in DRAFT status) with transaction support
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

	// Check if PR is deleted
	if pr.IsDeleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update deleted PR"})
		return
	}

	// Only allow updating DRAFT PRs
	if pr.Status != models.PRStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only DRAFT PRs can be updated"})
		return
	}

	// If items are being updated, validate before transaction
	if len(req.Items) > 0 {
		// Validate items without inventory checks
		tempPR := models.PurchaseRequest{ID: pr.ID}
		prItemsToCreate, itemCheckDetails, err := CheckAndCreatePRItems(&tempPR, req.Items)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                   err.Error(),
				"message":                 "PR update failed - validation error on items",
				"item_validation_summary": itemCheckDetails,
			})
			return
		}

		// Use transaction to ensure department update and items are updated together
		var updatedPR models.PurchaseRequest
		err = config.DB.Transaction(func(tx *gorm.DB) error {
			// Update department and purpose
			if req.Department != "" {
				pr.Department = req.Department
			}
			if req.Purpose != "" {
				pr.Purpose = req.Purpose
			}
			if err := tx.Save(&pr).Error; err != nil {
				return err
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PR and items: " + err.Error()})
			return
		}

		// Reload PR with items
		config.DB.Preload("Items").First(&updatedPR, pr.ID)

		c.JSON(http.StatusOK, gin.H{
			"message":                 "PR updated successfully",
			"data":                    updatedPR,
			"item_validation_summary": itemCheckDetails,
			"pr_items_created_count":  len(prItemsToCreate),
		})
	} else {
		// Update department and purpose only
		if req.Department != "" {
			pr.Department = req.Department
		}
		if req.Purpose != "" {
			pr.Purpose = req.Purpose
		}
		config.DB.Save(&pr)

		// Reload PR with items
		config.DB.Preload("Items").First(&pr, pr.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "PR updated successfully",
			"data":    pr,
		})
	}
}

// GetPR retrieves a PR by ID with role-based access control
func GetPR(c *gin.Context) {
	prID := c.Param("id")
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Check role-based access control
	role := userRole.(string)
	if role == "Admin" {
		// Admin can view all PRs
	} else if role == "Employee" {
		// Employee can only view own PRs
		if pr.RequesterID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied - can only view own PRs"})
			return
		}
	} else if role == "Manager" || role == "PurchaseOfficer" || role == "Executive" {
		// Manager/PurchaseOfficer/Executive can view non-DRAFT PRs
		if pr.Status == models.PRStatusDraft {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied - cannot view DRAFT PRs"})
			return
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, pr)
}

// GetPRList retrieves PRs based on role-based access control
func GetPRList(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	status := c.Query("status")

	var prs []models.PurchaseRequest
	query := config.DB

	role := userRole.(string)
	if role == "Admin" {
		// Admin can see all PRs
	} else if role == "Employee" {
		// Employee can only see own PRs
		query = query.Where("requester_id = ?", userID.(uint))
	} else if role == "Manager" || role == "PurchaseOfficer" || role == "Executive" {
		// Manager/PurchaseOfficer/Executive can see non-DRAFT PRs
		query = query.Where("status != ?", models.PRStatusDraft)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Don't show deleted PRs
	query = query.Where("is_deleted = ?", false)

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
func SubmitPR(c *gin.Context) {
	prID := c.Param("id")

	var pr models.PurchaseRequest
	if err := config.DB.Preload("Items").First(&pr, prID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Check if PR is deleted
	if pr.IsDeleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot submit deleted PR"})
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

	// Step 2: Change Status to PENDING
	pr.Status = models.PRStatusPending
	// Generate workflow ID with timestamp
	pr.WorkflowID = "WF_" + strconv.FormatUint(uint64(pr.ID), 10) + "_" + time.Now().Format("20060102150405")

	if err := config.DB.Save(&pr).Error; err != nil {
		log.Printf("failed to update PR status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PR status"})
		return
	}

	log.Printf("Updated PR %d status to PENDING with WorkflowID: %s", pr.ID, pr.WorkflowID)

	// Step 3: Trigger Approval by publishing event
	event := messaging.PRReadyForApprovalEvent{
		PRID:        pr.ID,
		PRNumber:    pr.PRNumber,
		RequesterID: pr.RequesterID,
		Department:  pr.Department,
		Purpose:     pr.Purpose,
		WorkflowID:  pr.WorkflowID,
		Timestamp:   time.Now(),
	}

	// Convert items with snapshot data
	for _, item := range pr.Items {
		event.Items = append(event.Items, messaging.PRItemPayload{
			SKU:          item.SKU,
			ItemName:     item.ItemName,
			Description:  item.Description,
			Quantity:     item.Quantity,
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
	log.Printf("Soft deleted PR %d", pr.ID)
	c.JSON(http.StatusOK, gin.H{"message": "PR deleted successfully"})
}
