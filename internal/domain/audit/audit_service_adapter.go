package audit

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// AuditServiceAdapter adapts the audit service to a consistent interface
type AuditServiceAdapter struct {
	// Internal services
	logAuditService    *LogAuditService
	getAuditLogsService *GetAuditLogsService
}

// NewAuditServiceAdapter creates a new AuditServiceAdapter
func NewAuditServiceAdapter(
	logAuditService *LogAuditService,
	getAuditLogsService *GetAuditLogsService,
) *AuditServiceAdapter {
	return &AuditServiceAdapter{
		logAuditService:    logAuditService,
		getAuditLogsService: getAuditLogsService,
	}
}

// LogAudit logs an audit event
func (a *AuditServiceAdapter) LogAudit(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	entityType string,
	entityID string,
	description string,
	details string,
	metadata map[string]interface{},
) (*entity.AuditLog, error) {
	// Create input for the service
	input := LogAuditInput{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		Details:     details,
		Metadata:    metadata,
	}

	// Log audit
	output, err := a.logAuditService.LogAudit(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.AuditLog{
		ID:          output.ID,
		UserID:      output.UserID,
		Action:      output.Action,
		EntityType:  output.EntityType,
		EntityID:    output.EntityID,
		Description: output.Description,
		Details:     output.Details,
		Metadata:    output.Metadata,
		PerformedBy: output.UserID,
		PerformedAt: output.CreatedAt,
		CreatedAt:   output.CreatedAt,
	}

	return result, nil
}

// GetAuditLogs gets a list of audit logs with pagination
func (a *AuditServiceAdapter) GetAuditLogs(
	ctx context.Context,
	page, pageSize int,
	entityType, entityID, action string,
	startDate, endDate time.Time,
) (*entity.PaginatedAuditLogs, error) {
	// Create input for the service
	input := GetAuditLogsInput{
		Page:       page,
		PageSize:   pageSize,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Get audit logs
	output, err := a.getAuditLogsService.GetAuditLogs(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.PaginatedAuditLogs{
		AuditLogs:  output.AuditLogs,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}

	return result, nil
}
