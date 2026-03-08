package handlers

import (
	"net/http"
	"strconv"

	"github.com/core-procurement/approval-service/services"
	"github.com/gin-gonic/gin"
)

// CreateApprovalRequest represents the request body for creating an approval instance
type CreateApprovalRequest struct {
	EntityType string `json:"entity_type" binding:"required"` // e.g., "PR"
	EntityID   uint   `json:"entity_id" binding:"required"`
	CreatedBy  uint   `json:"created_by" binding:"required"`
	WorkflowID string `json:"workflow_id" binding:"required"` // Correlation ID from Purchase Service
}

// ApprovalActionRequest represents the request body for approving/rejecting
type ApprovalActionRequest struct {
	Comment string `json:"comment"`
}

// CreateApproval creates a new approval instance for an entity
func CreateApproval(c *gin.Context) {
	var req CreateApprovalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	instance, err := services.CreateApprovalInstance(
		req.EntityType,
		req.EntityID,
		req.CreatedBy,
		req.WorkflowID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, instance)
}

// GetApproval retrieves approval status and steps by entity type and entity ID
func GetApproval(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_id"})
		return
	}

	instance, err := services.GetApprovalInstance(entityType, uint(entityID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// ApproveStep approves the current step and moves to the next one
func ApproveStep(c *gin.Context) {
	instanceIDStr := c.Param("id")

	instanceID, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ApprovalActionRequest

	// Bind JSON body - allow empty body
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and role from JWT context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in token"})
		return
	}

	instance, err := services.ApproveStep(uint(instanceID), userID.(uint), userRole.(string), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// RejectStep rejects the current step and marks instance as rejected
func RejectStep(c *gin.Context) {
	instanceIDStr := c.Param("id")

	instanceID, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ApprovalActionRequest

	// Bind JSON body - allow empty body
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and role from JWT context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in token"})
		return
	}

	instance, err := services.RejectStep(uint(instanceID), userID.(uint), userRole.(string), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// GetApprovalByWorkflow retrieves approval status and steps by workflow ID
func GetApprovalByWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	instance, err := services.GetApprovalInstanceByWorkflowID(workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval workflow not found"})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// ApproveStepByWorkflow approves the current step using workflow ID
func ApproveStepByWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	var req ApprovalActionRequest

	// Bind JSON body - allow empty body
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and role from JWT context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in token"})
		return
	}

	instance, err := services.ApproveStepByWorkflowID(workflowID, userID.(uint), userRole.(string), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// RejectStepByWorkflow rejects the current step using workflow ID
func RejectStepByWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	var req ApprovalActionRequest

	// Bind JSON body - allow empty body
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and role from JWT context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in token"})
		return
	}

	instance, err := services.RejectStepByWorkflowID(workflowID, userID.(uint), userRole.(string), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}
