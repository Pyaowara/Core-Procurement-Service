package models

import "time"

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

type ApprovalRole string

const (
	ApprovalRolePRCreator      ApprovalRole = "Employee"
	ApprovalRoleDepartmentHead ApprovalRole = "Manager"
	ApprovalRoleProcurement    ApprovalRole = "PurchaseOfficer"
	ApprovalRoleExecutive      ApprovalRole = "EXECUTIVE"
)

const (
	ActionApproved = "APPROVED"
	ActionRejected = "REJECTED"
)

// ApprovalInstance represents a workflow instance for an entity (e.g., PR)
type ApprovalInstance struct {
	ID          uint             `gorm:"primaryKey"`
	EntityType  string           `gorm:"index;not null"`       // e.g., "PR"
	EntityID    uint             `gorm:"index;not null"`       // e.g., PR ID
	WorkflowID  string           `gorm:"uniqueIndex;not null"` // Correlation ID from Purchase Service
	Status      ApprovalStatus   `gorm:"type:varchar(20);default:'PENDING'"`
	CurrentStep int              `gorm:"default:1"`
	CreatedBy   uint             `gorm:"not null"` // User who created the entity
	Steps       []ApprovalStep   `gorm:"foreignKey:InstanceID"`
	Actions     []ApprovalAction `gorm:"foreignKey:InstanceID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ApprovalStep represents a single approval step in the workflow
type ApprovalStep struct {
	ID         uint           `gorm:"primaryKey"`
	InstanceID uint           `gorm:"index;not null"`
	StepOrder  int            `gorm:"not null"`  // 1, 2, 3, 4
	ApproverID uint           `gorm:"default:0"` // Optional: Specific user ID (0 = any user with role)
	Role       ApprovalRole   `gorm:"type:varchar(50);not null"`
	Status     ApprovalStatus `gorm:"type:varchar(20);default:'PENDING'"`
	ActionAt   *time.Time     // When the action was taken
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ApprovalAction records every action taken on an approval instance
type ApprovalAction struct {
	ID         uint   `gorm:"primaryKey"`
	InstanceID uint   `gorm:"index;not null"`
	StepID     uint   `gorm:"index;not null"`
	ActorID    uint   `gorm:"not null"`                  // User who took the action
	ActionType string `gorm:"type:varchar(20);not null"` // APPROVED, REJECTED
	Comment    string `gorm:"type:text"`
	CreatedAt  time.Time
}
