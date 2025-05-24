package audit

import (
	"time"
	
	"github.com/google/uuid"
)

// CreateAuditLogService defines the interface for creating audit logs
type CreateAuditLogService interface {
	CreateAuditLog(input CreateAuditLogInput) (*CreateAuditLogOutput, error)
}

// CreateAuditLogInput represents the input for creating an audit log
type CreateAuditLogInput struct {
	Action      string
	Entity      string
	EntityID    uuid.UUID
	Metadata    map[string]interface{}
	PerformedBy uuid.UUID
}

// CreateAuditLogOutput represents the output of creating an audit log
type CreateAuditLogOutput struct {
	ID          uuid.UUID
	Action      string
	Entity      string
	EntityID    uuid.UUID
	Metadata    map[string]interface{}
	PerformedBy uuid.UUID
	CreatedAt   time.Time
}

// GetAuditLogsService defines the interface for getting audit logs
type GetAuditLogsService interface {
	GetAuditLogs(input GetAuditLogsInput) (*GetAuditLogsOutput, error)
}

// GetAuditLogsInput represents the input for getting audit logs
type GetAuditLogsInput struct {
	Page        int
	PageSize    int
	EntityType  string // Changed from Entity to match domain layer
	EntityID    *uuid.UUID
	Action      string
	PerformedBy *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
}

// GetAuditLogsOutput represents the output of getting audit logs
type GetAuditLogsOutput struct {
	AuditLogs   []AuditLogOutput
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

// AuditLogOutput represents an audit log output
type AuditLogOutput struct {
	ID          uuid.UUID
	Action      string
	Entity      string
	EntityID    uuid.UUID
	Metadata    map[string]interface{}
	PerformedBy uuid.UUID
	CreatedAt   time.Time
}
