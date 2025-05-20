package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	auditDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GetDataUploadAuditsService provides functionality for retrieving data upload audit logs
type GetDataUploadAuditsService struct {
	auditRepository auditDomain.AuditRepository
	uploadAuditRepository participantDomain.UploadAuditRepository
}

// NewGetDataUploadAuditsService creates a new GetDataUploadAuditsService
func NewGetDataUploadAuditsService(
	auditRepository auditDomain.AuditRepository,
	uploadAuditRepository participantDomain.UploadAuditRepository,
) *GetDataUploadAuditsService {
	return &GetDataUploadAuditsService{
		auditRepository: auditRepository,
		uploadAuditRepository: uploadAuditRepository,
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

	// Get upload audits from repository
	uploadAudits, totalCount, err := s.uploadAuditRepository.List(input.Page, input.PageSize)
	if err != nil {
		return GetDataUploadAuditsOutput{}, fmt.Errorf("failed to list upload audits: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	// Map domain upload audits to service data upload audits
	dataUploadAudits := make([]DataUploadAudit, 0, len(uploadAudits))
	for _, audit := range uploadAudits {
		dataUploadAudits = append(dataUploadAudits, DataUploadAudit{
			ID:                  audit.ID,
			UploadedBy:          audit.UploadedBy,
			UploadedAt:          audit.UploadDate,
			FileName:            audit.FileName,
			TotalUploaded:       audit.TotalRows,
			SuccessfullyImported: audit.SuccessfulRows,
			DuplicatesSkipped:   audit.TotalRows - audit.SuccessfulRows - audit.ErrorCount,
			ErrorsEncountered:   audit.ErrorCount,
			Status:              audit.Status,
			Details:             "",
			OperationType:       "UPLOAD_PARTICIPANTS",
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
