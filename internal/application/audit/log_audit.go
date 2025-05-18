package audit

import (
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// LogAuditService provides functionality for logging audit events
type LogAuditService struct {
	auditRepository audit.AuditRepository
}

// NewLogAuditService creates a new LogAuditService
func NewLogAuditService(auditRepository audit.AuditRepository) *LogAuditService {
	return &LogAuditService{
		auditRepository: auditRepository,
	}
}

// LogAudit logs an audit event
func (s *LogAuditService) LogAudit(
	action string,
	entityType string,
	entityID uuid.UUID,
	userID uuid.UUID,
	summary string,
	details string,
) error {
	if action == "" {
		return fmt.Errorf("action is required")
	}
	
	if entityType == "" {
		return fmt.Errorf("entity type is required")
	}
	
	if entityID == uuid.Nil {
		return fmt.Errorf("entity ID is required")
	}
	
	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	
	auditLog := &audit.AuditLog{
		ID:         uuid.New(),
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID.String(),
		UserID:     userID,
		Description: summary,
		Metadata:    make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}
	
	// Add details to metadata
	auditLog.Metadata["details"] = details
	
	if err := s.auditRepository.Create(auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	
	return nil
}

// AuditService provides a simplified interface for logging audit events
type AuditService struct {
	logAuditService *LogAuditService
}

// NewAuditService creates a new AuditService
func NewAuditService(logAuditService *LogAuditService) *AuditService {
	return &AuditService{
		logAuditService: logAuditService,
	}
}

// LogAudit logs an audit event
func (s *AuditService) LogAudit(
	action string,
	entityType string,
	entityID uuid.UUID,
	userID uuid.UUID,
	summary string,
	details string,
) error {
	return s.logAuditService.LogAudit(action, entityType, entityID, userID, summary, details)
}
