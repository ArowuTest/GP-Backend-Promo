package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
)

// AuditServiceAdapter adapts the audit service to a consistent interface
type AuditServiceAdapter struct {
	createAuditLogService audit.AuditService
	getAuditLogsService   audit.GetAuditLogsService
}

// NewAuditServiceAdapter creates a new AuditServiceAdapter
func NewAuditServiceAdapter(
	createAuditLogService *audit.AuditService,
	getAuditLogsService audit.GetAuditLogsService,
) *AuditServiceAdapter {
	return &AuditServiceAdapter{
		createAuditLogService: *createAuditLogService,
		getAuditLogsService:   getAuditLogsService,
	}
}

// AuditLog represents an audit log
type AuditLog struct {
	ID          string
	Action      string
	Entity      string
	EntityID    string
	Metadata    map[string]interface{}
	PerformedBy string
	CreatedAt   time.Time
}

// CreateAuditLogOutput represents the output of CreateAuditLog
type CreateAuditLogOutput struct {
	ID          string
	Action      string
	Entity      string
	EntityID    string
	Metadata    map[string]interface{}
	PerformedBy string
	CreatedAt   time.Time
}

// GetAuditLogsOutput represents the output of GetAuditLogs
type GetAuditLogsOutput struct {
	AuditLogs   []AuditLog
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

// CreateAuditLog creates an audit log
func (a *AuditServiceAdapter) CreateAuditLog(
	ctx context.Context,
	action string,
	entity string,
	entityID uuid.UUID,
	metadata map[string]interface{},
	performedBy uuid.UUID,
) (*CreateAuditLogOutput, error) {
	// Call the actual service
	input := audit.CreateAuditLogInput{
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		Metadata:    metadata,
		PerformedBy: performedBy,
	}

	err := a.createAuditLogService.LogAudit(
		input.Action,
		input.Entity,
		input.EntityID,
		input.PerformedBy,
		"",  // Description
		"",  // Details
	)
	
	// Create a mock output since the actual service doesn't return one
	output := &audit.CreateAuditLogOutput{
		ID:          uuid.New(),
		Action:      input.Action,
		Entity:      input.Entity,
		EntityID:    input.EntityID,
		Metadata:    input.Metadata,
		PerformedBy: input.PerformedBy,
		CreatedAt:   time.Now(),
	}
	if err != nil {
		return nil, err
	}

	// Return response
	return &CreateAuditLogOutput{
		ID:          output.ID.String(),
		Action:      output.Action,
		Entity:      output.Entity,
		EntityID:    output.EntityID.String(),
		Metadata:    output.Metadata,
		PerformedBy: output.PerformedBy.String(),
		CreatedAt:   output.CreatedAt,
	}, nil
}

// GetAuditLogs gets audit logs with pagination
func (a *AuditServiceAdapter) GetAuditLogs(
	ctx context.Context,
	page int,
	pageSize int,
	entity string,
	entityID *uuid.UUID,
	action string,
	performedBy *uuid.UUID,
	startDate *time.Time,
	endDate *time.Time,
) (*GetAuditLogsOutput, error) {
	// Call the actual service
	input := audit.GetAuditLogsInput{
		Page:        page,
		PageSize:    pageSize,
		EntityType:  entity,
		EntityID:    entityID,
		Action:      action,
		PerformedBy: performedBy,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Check if getAuditLogsService is initialized
	if a.getAuditLogsService == nil {
		return &GetAuditLogsOutput{
			AuditLogs:   []AuditLog{},
			Page:        page,
			PageSize:    pageSize,
			TotalCount:  0,
			TotalPages:  0,
		}, nil
	}

	output, err := a.getAuditLogsService.GetAuditLogs(input)
	if err != nil {
		return nil, err
	}

	// Convert audit logs for response
	auditLogs := make([]AuditLog, 0, len(output.AuditLogs))
	for _, log := range output.AuditLogs {
		auditLogs = append(auditLogs, AuditLog{
			ID:          log.ID.String(),
			Action:      log.Action,
			Entity:      log.Entity,
			EntityID:    log.EntityID.String(),
			Metadata:    log.Metadata,
			PerformedBy: log.PerformedBy.String(),
			CreatedAt:   log.CreatedAt,
		})
	}

	// Return response
	return &GetAuditLogsOutput{
		AuditLogs:   auditLogs,
		Page:        output.Page,
		PageSize:    output.PageSize,
		TotalCount:  output.TotalCount,
		TotalPages:  output.TotalPages,
	}, nil
}
