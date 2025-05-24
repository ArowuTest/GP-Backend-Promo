package audit

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GetAuditLogsServiceImpl provides functionality for retrieving audit logs
type GetAuditLogsServiceImpl struct {
	auditRepository audit.AuditRepository
}

// NewGetAuditLogsService creates a new GetAuditLogsService
func NewGetAuditLogsService(auditRepository audit.AuditRepository) *GetAuditLogsServiceImpl {
	return &GetAuditLogsServiceImpl{
		auditRepository: auditRepository,
	}
}

// AuditLogFilters defines the filters for retrieving audit logs
type AuditLogFilters struct {
	Action      string
	EntityType  string
	EntityID    uuid.UUID
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
}

// GetAuditLogs retrieves a paginated list of audit logs
func (s *GetAuditLogsServiceImpl) GetAuditLogs(ctx context.Context, input GetAuditLogsInput) (*GetAuditLogsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	
	// Convert to domain filters
	filters := audit.AuditLogFilters{
		Action:     input.Action,
		EntityType: input.EntityType,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}
	
	if input.StartDate != nil {
		filters.StartDate = *input.StartDate
	}
	
	if input.EndDate != nil {
		filters.EndDate = *input.EndDate
	}
	
	if input.EntityID != nil {
		// Skip setting EntityID if it's not defined in domain struct
	}
	
	if input.PerformedBy != nil {
		filters.UserID = *input.PerformedBy
	}
	
	auditLogs, totalCount, err := s.auditRepository.List(filters, input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}
	
	// Convert domain audit logs to application output
	auditLogOutputs := make([]AuditLogOutput, 0, len(auditLogs))
	for _, log := range auditLogs {
		auditLogOutputs = append(auditLogOutputs, AuditLogOutput{
			ID:          log.ID,
			Action:      log.Action,
			Entity:      log.EntityType,
			EntityID:    uuid.MustParse(log.EntityID),
			Metadata:    log.Metadata,
			PerformedBy: log.UserID,
			CreatedAt:   log.CreatedAt,
		})
	}
	
	return &GetAuditLogsOutput{
		AuditLogs:  auditLogOutputs,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
