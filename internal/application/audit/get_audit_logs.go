package audit

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GetAuditLogsInput represents the input for the GetAuditLogs use case
type GetAuditLogsInput struct {
	StartDate  time.Time
	EndDate    time.Time
	ActionType string
	UserID     string
	Page       int
	PageSize   int
}

// GetAuditLogsOutput represents the output from the GetAuditLogs use case
type GetAuditLogsOutput struct {
	AuditLogs []audit.AuditLog
	Total     int64
	Page      int
	PageSize  int
}

// GetAuditLogsUseCase defines the use case for retrieving audit logs
type GetAuditLogsUseCase struct {
	auditRepo audit.Repository
}

// NewGetAuditLogsUseCase creates a new GetAuditLogsUseCase
func NewGetAuditLogsUseCase(auditRepo audit.Repository) *GetAuditLogsUseCase {
	return &GetAuditLogsUseCase{
		auditRepo: auditRepo,
	}
}

// Execute performs the get audit logs use case
func (uc *GetAuditLogsUseCase) Execute(ctx context.Context, input GetAuditLogsInput) (GetAuditLogsOutput, error) {
	// Set default page size if not provided
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	// Set default page if not provided
	if input.Page <= 0 {
		input.Page = 1
	}

	// Prepare filter criteria
	filter := audit.AuditLogFilter{
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		ActionType: input.ActionType,
		UserID:     input.UserID,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}

	// Get audit logs from repository
	auditLogs, err := uc.auditRepo.GetAuditLogs(ctx, filter)
	if err != nil {
		return GetAuditLogsOutput{}, err
	}

	// Get total count for pagination
	total, err := uc.auditRepo.CountAuditLogs(ctx, filter)
	if err != nil {
		return GetAuditLogsOutput{}, err
	}

	return GetAuditLogsOutput{
		AuditLogs: auditLogs,
		Total:     total,
		Page:      input.Page,
		PageSize:  input.PageSize,
	}, nil
}
