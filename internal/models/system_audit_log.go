package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemAuditLog represents a system-wide audit log entry for tracking admin actions
type SystemAuditLog struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid" json:"user_id"`
	User          AdminUser      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ActionType    string         `gorm:"not null" json:"action_type"` // e.g., "LOGIN", "CREATE_USER", "UPDATE_PRIZE_STRUCTURE", "EXECUTE_DRAW"
	ResourceType  string         `gorm:"not null" json:"resource_type"` // e.g., "USER", "PRIZE_STRUCTURE", "DRAW"
	ResourceID    string         `json:"resource_id,omitempty"`
	Description   string         `gorm:"not null" json:"description"`
	IPAddress     string         `json:"ip_address,omitempty"`
	UserAgent     string         `json:"user_agent,omitempty"`
	ActionDetails string         `gorm:"type:text" json:"action_details,omitempty"` // JSON string with detailed information
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (a *SystemAuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// CreateSystemAuditLog creates a new system audit log entry
func CreateSystemAuditLog(db *gorm.DB, userID uuid.UUID, actionType, resourceType, resourceID, description, ipAddress, userAgent, actionDetails string) error {
	auditLog := SystemAuditLog{
		UserID:        userID,
		ActionType:    actionType,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Description:   description,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		ActionDetails: actionDetails,
	}
	
	return db.Create(&auditLog).Error
}
