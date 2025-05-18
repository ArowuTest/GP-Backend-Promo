package audit

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GetAuditLogsService provides functionality for retrieving audit logs
type GetAuditLogsService struct {
	auditRepository audit.AuditRepository
}

// NewGetAuditLogsService creates a new GetAuditLogsService
func NewGetAuditLogsService(auditRepository audit.AuditRepository) *GetAuditLogsService {
	return &GetAuditLogsService{
		auditRepository: auditRepository,
	}
}

// GetAuditLogsInput defines the input for the GetAuditLogs use case
type GetAuditLogsInput struct {
	Page     int
	PageSize int
	Filters  AuditLogFilters
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

// GetAuditLogsOutput defines the output for the GetAuditLogs use case
type GetAuditLogsOutput struct {
	AuditLogs   []audit.AuditLog
	TotalCount  int
	Page        int
	PageSize    int
	TotalPages  int
}

// GetAuditLogs retrieves a paginated list of audit logs
func (s *GetAuditLogsService) GetAuditLogs(ctx context.Context, input GetAuditLogsInput) (*GetAuditLogsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	
	// Convert to domain filters
	filters := audit.AuditLogFilters{
		Action:     input.Filters.Action,
		EntityType: input.Filters.EntityType,
		UserID:     input.Filters.UserID,
		StartDate:  input.Filters.StartDate,
		EndDate:    input.Filters.EndDate,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}
	
	auditLogs, totalCount, err := s.auditRepository.List(filters, input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}
	
	return &GetAuditLogsOutput{
		AuditLogs:  auditLogs,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
