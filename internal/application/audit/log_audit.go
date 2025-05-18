package application

import (
	"time"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// LogAuditUseCase represents the use case for logging an audit event
type LogAuditUseCase struct {
	auditRepository audit.AuditRepository
}

// NewLogAuditUseCase creates a new LogAuditUseCase
func NewLogAuditUseCase(
	auditRepository audit.AuditRepository,
) *LogAuditUseCase {
	return &LogAuditUseCase{
		auditRepository: auditRepository,
	}
}

// LogAuditInput represents the input for the log audit use case
type LogAuditInput struct {
	UserID      uuid.UUID
	Username    string
	Action      string
	EntityType  string
	EntityID    string
	Description string
	IPAddress   string
	UserAgent   string
	Metadata    map[string]interface{}
}

// LogAuditOutput represents the output of the log audit use case
type LogAuditOutput struct {
	AuditLog *audit.AuditLog
}

// Execute logs an audit event
func (uc *LogAuditUseCase) Execute(input LogAuditInput) (*LogAuditOutput, error) {
	// Create audit log entity
	auditLog := &audit.AuditLog{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Username:    input.Username,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		Description: input.Description,
		IPAddress:   input.IPAddress,
		UserAgent:   input.UserAgent,
		Metadata:    input.Metadata,
		CreatedAt:   time.Now(),
	}

	// Validate audit log
	if err := audit.ValidateAuditLog(auditLog); err != nil {
		return nil, audit.NewAuditError(audit.ErrInvalidAuditLog, "Invalid audit log", err)
	}

	// Save audit log
	if err := uc.auditRepository.CreateAuditLog(auditLog); err != nil {
		return nil, audit.NewAuditError("AUDIT_LOG_CREATION_FAILED", "Failed to create audit log", err)
	}

	return &LogAuditOutput{
		AuditLog: auditLog,
	}, nil
}
