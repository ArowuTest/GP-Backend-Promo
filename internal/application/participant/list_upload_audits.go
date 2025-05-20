package participant

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// ListUploadAuditsService provides functionality for retrieving upload audit logs
type ListUploadAuditsService struct {
	participantRepository participantDomain.ParticipantRepository
}

// NewListUploadAuditsService creates a new ListUploadAuditsService
func NewListUploadAuditsService(participantRepository participantDomain.ParticipantRepository) *ListUploadAuditsService {
	return &ListUploadAuditsService{
		participantRepository: participantRepository,
	}
}

// ListUploadAuditsInput defines the input for the ListUploadAudits use case
type ListUploadAuditsInput struct {
	Page     int
	PageSize int
}

// UploadAudit represents an upload audit record
type UploadAudit struct {
	ID             uuid.UUID
	UploadedBy     uuid.UUID
	UploadDate     time.Time
	FileName       string
	Status         string
	TotalRows      int
	SuccessfulRows int
	ErrorCount     int
}

// ListUploadAuditsOutput defines the output for the ListUploadAudits use case
type ListUploadAuditsOutput struct {
	UploadAudits []UploadAudit
	TotalCount   int
	Page         int
	PageSize     int
	TotalPages   int
}

// ListUploadAudits retrieves upload audit logs
func (s *ListUploadAuditsService) ListUploadAudits(ctx context.Context, input ListUploadAuditsInput) (ListUploadAuditsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}

	if input.PageSize < 1 {
		input.PageSize = 10
	}

	// Get upload audits from repository
	audits, totalCount, err := s.participantRepository.ListUploadAudits(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListUploadAuditsOutput{}, fmt.Errorf("failed to list upload audits: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	// Map domain audits to service audits
	uploadAudits := make([]UploadAudit, 0, len(audits))
	for _, audit := range audits {
		uploadAudits = append(uploadAudits, UploadAudit{
			ID:             audit.ID,
			UploadedBy:     audit.UploadedBy,
			UploadDate:     audit.UploadDate,
			FileName:       audit.FileName,
			Status:         audit.Status,
			TotalRows:      audit.TotalRows,
			SuccessfulRows: audit.SuccessfulRows,
			ErrorCount:     audit.ErrorCount,
		})
	}

	return ListUploadAuditsOutput{
		UploadAudits: uploadAudits,
		TotalCount:   totalCount,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalPages:   totalPages,
	}, nil
}
