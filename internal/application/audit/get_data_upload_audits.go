package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	auditDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GetDataUploadAuditsService provides functionality for retrieving data upload audit logs
type GetDataUploadAuditsService struct {
	auditRepository auditDomain.AuditRepository
}

// NewGetDataUploadAuditsService creates a new GetDataUploadAuditsService
func NewGetDataUploadAuditsService(auditRepository auditDomain.AuditRepository) *GetDataUploadAuditsService {
	return &GetDataUploadAuditsService{
		auditRepository: auditRepository,
	}
}

// GetDataUploadAuditsInput defines the input for the GetDataUploadAudits use case
type GetDataUploadAuditsInput struct {
	Page     int
	PageSize int
}

// DataUploadAudit represents a data upload audit record
type DataUploadAudit struct {
	ID                  uuid.UUID
	UploadedBy          uuid.UUID
	UploadedAt          time.Time
	FileName            string
	TotalUploaded       int
	SuccessfullyImported int
	DuplicatesSkipped   int
	ErrorsEncountered   int
	Status              string
	Details             string
	OperationType       string
}

// GetDataUploadAuditsOutput defines the output for the GetDataUploadAudits use case
type GetDataUploadAuditsOutput struct {
	DataUploadAudits []DataUploadAudit
	TotalCount       int
	Page             int
	PageSize         int
	TotalPages       int
}

// GetDataUploadAudits retrieves data upload audit logs
func (s *GetDataUploadAuditsService) GetDataUploadAudits(ctx context.Context, input GetDataUploadAuditsInput) (GetDataUploadAuditsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}

	if input.PageSize < 1 {
		input.PageSize = 10
	}

	// Get data upload audits from repository
	// This would typically filter audits related to data uploads
	filters := auditDomain.AuditLogFilters{
		Action: "UPLOAD_PARTICIPANTS",
		Page:   input.Page,
		PageSize: input.PageSize,
	}

	auditLogs, totalCount, err := s.auditRepository.List(filters, input.Page, input.PageSize)
	if err != nil {
		return GetDataUploadAuditsOutput{}, fmt.Errorf("failed to list data upload audits: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	// Map domain audit logs to data upload audits
	dataUploadAudits := make([]DataUploadAudit, 0, len(auditLogs))
	for _, log := range auditLogs {
		dataUploadAudits = append(dataUploadAudits, DataUploadAudit{
			ID:                  log.ID,
			UploadedBy:          log.UserID,
			UploadedAt:          log.CreatedAt,
			FileName:            log.Details,
			TotalUploaded:       log.TotalCount,
			SuccessfullyImported: log.SuccessCount,
			DuplicatesSkipped:   log.DuplicateCount,
			ErrorsEncountered:   log.ErrorCount,
			Status:              log.Status,
			Details:             log.Summary,
			OperationType:       log.Action,
		})
	}

	return GetDataUploadAuditsOutput{
		DataUploadAudits: dataUploadAudits,
		TotalCount:       totalCount,
		Page:             input.Page,
		PageSize:         input.PageSize,
		TotalPages:       totalPages,
	}, nil
}
