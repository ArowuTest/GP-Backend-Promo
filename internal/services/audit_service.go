package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService handles audit logging operations
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new AuditService with the provided database connection
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// ActionType represents the type of action being audited
type ActionType string

// ResourceType represents the type of resource being acted upon
type ResourceType string

// Common action types
const (
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionView   ActionType = "VIEW"
	ActionLogin  ActionType = "LOGIN"
	ActionLogout ActionType = "LOGOUT"
)

// Common resource types
const (
	ResourceUser           ResourceType = "USER"
	ResourcePrizeStructure ResourceType = "PRIZE_STRUCTURE"
	ResourceDraw           ResourceType = "DRAW"
	ResourceWinner         ResourceType = "WINNER"
	ResourceParticipant    ResourceType = "PARTICIPANT"
)

// CreateAuditLog creates a new audit log entry
func (s *AuditService) CreateAuditLog(log *models.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	
	return s.db.Create(log).Error
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(
	page, pageSize int,
	startDate, endDate *time.Time,
	userID, actionType, resourceType, resourceID string,
) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64
	
	query := s.db.Model(&models.AuditLog{})
	
	// Apply filters
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}
	
	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			query = query.Where("user_id = ?", userUUID)
		}
	}
	
	if actionType != "" {
		query = query.Where("action = ?", actionType)
	}
	
	if resourceType != "" {
		query = query.Where("entity_type = ?", resourceType)
	}
	
	if resourceID != "" {
		resourceUUID, err := uuid.Parse(resourceID)
		if err == nil {
			query = query.Where("entity_id = ?", resourceUUID)
		}
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// LogUserAction creates an audit log for a user action
func (s *AuditService) LogUserAction(
	userID string,
	actionType ActionType,
	resourceType ResourceType,
	resourceID string,
	description string,
	ipAddress string,
	userAgent string,
	details interface{},
) error {
	// Convert details to JSON
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal action details: %w", err)
		}
		detailsJSON = string(detailsBytes)
	}
	
	// Parse UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	
	var entityUUID uuid.UUID
	if resourceID != "" {
		entityUUID, err = uuid.Parse(resourceID)
		if err != nil {
			return fmt.Errorf("invalid entity ID: %w", err)
		}
	} else {
		entityUUID = uuid.Nil
	}
	
	// Create audit log
	log := &models.AuditLog{
		ID:        uuid.New(),
		UserID:    userUUID,
		Action:    string(actionType),
		EntityType: string(resourceType),
		EntityID:  entityUUID,
		Details:   detailsJSON,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}
	
	return s.CreateAuditLog(log)
}

// GetAuditLogTypes retrieves all unique action types and resource types
func (s *AuditService) GetAuditLogTypes() ([]string, []string, error) {
	var actionTypes []string
	var resourceTypes []string
	
	// Get unique action types
	if err := s.db.Model(&models.AuditLog{}).
		Distinct("action").
		Pluck("action", &actionTypes).Error; err != nil {
		return nil, nil, err
	}
	
	// Get unique resource types
	if err := s.db.Model(&models.AuditLog{}).
		Distinct("entity_type").
		Pluck("entity_type", &resourceTypes).Error; err != nil {
		return nil, nil, err
	}
	
	return actionTypes, resourceTypes, nil
}
