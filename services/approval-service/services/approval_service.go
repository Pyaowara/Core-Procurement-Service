package services

import (
	"fmt"
	"log"
	"time"

	"github.com/core-procurement/approval-service/config"
	"github.com/core-procurement/approval-service/messaging"
	"github.com/core-procurement/approval-service/models"
)

// CreateApprovalInstance creates a new approval workflow for an entity (e.g., PR)
// It creates one ApprovalInstance and four ApprovalStep records
// Approvals are role-based: any user with the required role can approve each step
func CreateApprovalInstance(entityType string, entityID uint, createdByUserID uint, workflowID string) (*models.ApprovalInstance, error) {

	instance := models.ApprovalInstance{
		EntityType:  entityType,
		EntityID:    entityID,
		WorkflowID:  workflowID,
		Status:      models.ApprovalStatusPending,
		CurrentStep: 2,
		CreatedBy:   createdByUserID,
	}

	if err := config.DB.Create(&instance).Error; err != nil {
		log.Printf("failed to create approval instance: %v", err)
		return nil, err
	}

	// Define approval steps in order
	steps := []struct {
		stepOrder int
		role      models.ApprovalRole
	}{
		{1, models.ApprovalRolePRCreator},
		{2, models.ApprovalRoleDepartmentHead},
		{3, models.ApprovalRoleProcurement},
		{4, models.ApprovalRoleExecutive},
	}

	for _, step := range steps {
		status := models.ApprovalStatusPending
		approverID := uint(0)
		var actionAt *time.Time

		// PR submit counts as sender approval for step 1.
		if step.stepOrder == 1 {
			now := time.Now()
			status = models.ApprovalStatusApproved
			approverID = createdByUserID
			actionAt = &now
		}

		approvalStep := models.ApprovalStep{
			InstanceID: instance.ID,
			StepOrder:  step.stepOrder,
			ApproverID: approverID,
			Role:       step.role,
			Status:     status,
			ActionAt:   actionAt,
		}

		if err := config.DB.Create(&approvalStep).Error; err != nil {
			log.Printf("failed to create approval step: %v", err)
			return nil, err
		}

		if step.stepOrder == 1 {
			action := models.ApprovalAction{
				InstanceID: instance.ID,
				StepID:     approvalStep.ID,
				ActorID:    createdByUserID,
				ActionType: models.ActionApproved,
				Comment:    "Auto-approved on PR submit by requester",
			}

			if err := config.DB.Create(&action).Error; err != nil {
				log.Printf("failed to record auto approval action: %v", err)
				return nil, err
			}
		}
	}

	// Reload instance with steps
	if err := config.DB.Preload("Steps").First(&instance).Error; err != nil {
		log.Printf("failed to reload approval instance: %v", err)
		return nil, err
	}

	return &instance, nil
}

// GetApprovalInstance retrieves an approval instance by entity type and entity ID
func GetApprovalInstance(entityType string, entityID uint) (*models.ApprovalInstance, error) {
	var instance models.ApprovalInstance

	if err := config.DB.Preload("Steps").Preload("Actions").
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		First(&instance).Error; err != nil {
		log.Printf("failed to get approval instance: %v", err)
		return nil, err
	}

	return &instance, nil
}

// GetApprovalInstanceByID retrieves an approval instance by ID
func GetApprovalInstanceByID(id uint) (*models.ApprovalInstance, error) {
	var instance models.ApprovalInstance

	if err := config.DB.Preload("Steps").Preload("Actions").
		First(&instance, id).Error; err != nil {
		log.Printf("failed to get approval instance: %v", err)
		return nil, err
	}

	return &instance, nil
}

// GetApprovalInstanceByWorkflowID retrieves an approval instance by workflow ID
func GetApprovalInstanceByWorkflowID(workflowID string) (*models.ApprovalInstance, error) {
	var instance models.ApprovalInstance

	if err := config.DB.Preload("Steps").Preload("Actions").
		Where("workflow_id = ?", workflowID).
		First(&instance).Error; err != nil {
		log.Printf("failed to get approval instance by workflow_id: %v", err)
		return nil, err
	}

	return &instance, nil
}

// VerifyApprovalRole checks if userRole matches the required approval role
func VerifyApprovalRole(userRole string, requiredApprovalRole models.ApprovalRole) bool {
	// Map user roles to approval roles
	// User roles: Employee, Manager, PurchaseOfficer, Executive, Admin
	// Approval roles: PR_CREATOR, DEPARTMENT_HEAD, PROCUREMENT, EXECUTIVE

	roleMapping := map[models.ApprovalRole][]string{
		models.ApprovalRolePRCreator:      {"Employee", "Manager", "PurchaseOfficer", "Executive", "Admin"}, // Anyone can be requester
		models.ApprovalRoleDepartmentHead: {"Manager", "Executive", "Admin"},                                // Managers and above
		models.ApprovalRoleProcurement:    {"PurchaseOfficer", "Executive", "Admin"},                        // Purchase officers and above
		models.ApprovalRoleExecutive:      {"Executive", "Admin"},                                           // Executives and admins
	}

	allowedRoles, exists := roleMapping[requiredApprovalRole]
	if !exists {
		log.Printf("unknown approval role: %s", requiredApprovalRole)
		return false
	}

	for _, role := range allowedRoles {
		if role == userRole {
			return true
		}
	}

	return false
}

// ApproveStep approves the current step and moves to the next one
func ApproveStep(instanceID uint, actorID uint, userRole string, comment string) (*models.ApprovalInstance, error) {
	instance, err := GetApprovalInstanceByID(instanceID)
	if err != nil {
		return nil, err
	}

	// Check if already completed
	if instance.Status != models.ApprovalStatusPending {
		return nil, fmt.Errorf("approval instance is already %s", instance.Status)
	}

	// Get current step
	var currentStep models.ApprovalStep
	if err := config.DB.Where("instance_id = ? AND step_order = ?", instanceID, instance.CurrentStep).
		First(&currentStep).Error; err != nil {
		log.Printf("failed to get current step: %v", err)
		return nil, err
	}

	// Verify user has required role for this step
	if !VerifyApprovalRole(userRole, currentStep.Role) {
		return nil, fmt.Errorf("user role '%s' cannot approve step %d (requires %s)", userRole, currentStep.StepOrder, currentStep.Role)
	}

	// Update step status
	now := models.ApprovalAction{}
	if err := config.DB.Model(&currentStep).
		Updates(map[string]interface{}{
			"status":    models.ApprovalStatusApproved,
			"action_at": now.CreatedAt,
		}).Error; err != nil {
		log.Printf("failed to update step status: %v", err)
		return nil, err
	}

	// Record action
	action := models.ApprovalAction{
		InstanceID: instanceID,
		StepID:     currentStep.ID,
		ActorID:    actorID,
		ActionType: string(models.ActionApproved),
		Comment:    comment,
	}

	if err := config.DB.Create(&action).Error; err != nil {
		log.Printf("failed to record approval action: %v", err)
		return nil, err
	}

	// Check if this is the last step
	var totalSteps int64
	config.DB.Model(&models.ApprovalStep{}).Where("instance_id = ?", instanceID).Count(&totalSteps)

	if instance.CurrentStep >= int(totalSteps) {
		// All steps approved, mark instance as approved
		if err := config.DB.Model(instance).Update("status", models.ApprovalStatusApproved).Error; err != nil {
			log.Printf("failed to update instance status: %v", err)
			return nil, err
		}

		// Publish approval completed event to RabbitMQ
		if instance.EntityType == "PR" {
			event := messaging.ApprovalCompletedEvent{
				PRID:       instance.EntityID,
				WorkflowID: instance.WorkflowID,
				Status:     "APPROVED",
				ApprovedAt: time.Now(),
			}
			eventBytes, _ := messaging.MarshalEvent(event)
			if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventApprovalCompleted, eventBytes); err != nil {
				log.Printf("failed to publish approval completed event: %v", err)
			} else {
				log.Printf("Published approval completed event for PR %d", instance.EntityID)
			}
		}
	} else {
		// Move to next step
		if err := config.DB.Model(instance).Update("current_step", instance.CurrentStep+1).Error; err != nil {
			log.Printf("failed to update current step: %v", err)
			return nil, err
		}
	}

	// Reload instance
	return GetApprovalInstanceByID(instanceID)
}

// RejectStep rejects the current step and marks instance as rejected
func RejectStep(instanceID uint, actorID uint, userRole string, comment string) (*models.ApprovalInstance, error) {
	instance, err := GetApprovalInstanceByID(instanceID)
	if err != nil {
		return nil, err
	}

	// Check if already completed
	if instance.Status != models.ApprovalStatusPending {
		return nil, fmt.Errorf("approval instance is already %s", instance.Status)
	}

	// Get current step
	var currentStep models.ApprovalStep
	if err := config.DB.Where("instance_id = ? AND step_order = ?", instanceID, instance.CurrentStep).
		First(&currentStep).Error; err != nil {
		log.Printf("failed to get current step: %v", err)
		return nil, err
	}

	// Verify user has required role for this step
	if !VerifyApprovalRole(userRole, currentStep.Role) {
		return nil, fmt.Errorf("user role '%s' cannot reject step %d (requires %s)", userRole, currentStep.StepOrder, currentStep.Role)
	}

	// Update step status
	now := models.ApprovalAction{}
	if err := config.DB.Model(&currentStep).
		Updates(map[string]interface{}{
			"status":    models.ApprovalStatusRejected,
			"action_at": now.CreatedAt,
		}).Error; err != nil {
		log.Printf("failed to update step status: %v", err)
		return nil, err
	}

	// Record action
	action := models.ApprovalAction{
		InstanceID: instanceID,
		StepID:     currentStep.ID,
		ActorID:    actorID,
		ActionType: string(models.ActionRejected),
		Comment:    comment,
	}

	if err := config.DB.Create(&action).Error; err != nil {
		log.Printf("failed to record rejection action: %v", err)
		return nil, err
	}

	// Mark instance as rejected
	if err := config.DB.Model(instance).Update("status", models.ApprovalStatusRejected).Error; err != nil {
		log.Printf("failed to update instance status: %v", err)
		return nil, err
	}

	// Publish approval rejected event to RabbitMQ
	if instance.EntityType == "PR" {
		event := messaging.ApprovalRejectedEvent{
			PRID:       instance.EntityID,
			WorkflowID: instance.WorkflowID,
			Reason:     comment,
			RejectedAt: time.Now(),
		}
		eventBytes, _ := messaging.MarshalEvent(event)
		if err := messaging.MQClient.PublishMessage(messaging.ExchangeName, messaging.EventApprovalRejected, eventBytes); err != nil {
			log.Printf("failed to publish approval rejected event: %v", err)
		} else {
			log.Printf("Published approval rejected event for PR %d", instance.EntityID)
		}
	}

	// Reload instance
	return GetApprovalInstanceByID(instanceID)
}

// ApproveStepByWorkflowID approves the current step using workflow ID
func ApproveStepByWorkflowID(workflowID string, actorID uint, userRole string, comment string) (*models.ApprovalInstance, error) {
	instance, err := GetApprovalInstanceByWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}
	return ApproveStep(instance.ID, actorID, userRole, comment)
}

// RejectStepByWorkflowID rejects the current step using workflow ID
func RejectStepByWorkflowID(workflowID string, actorID uint, userRole string, comment string) (*models.ApprovalInstance, error) {
	instance, err := GetApprovalInstanceByWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}
	return RejectStep(instance.ID, actorID, userRole, comment)
}
