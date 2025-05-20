package participant

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// ListUploadAuditsService provides functionality for retrieving upload audit logs
type ListUploadAuditsService struct {
	uploadAuditRepository participantDomain.UploadAuditRepository
}

// NewListUploadAuditsService creates a new ListUploadAuditsService
func NewListUploadAuditsService(uploadAuditRepository participantDomain.UploadAuditRepository) *ListUploadAuditsService {
	return &ListUploadAuditsService{
		uploadAuditRepository: uploadAuditRepository,
	}
}

// ListUploadAuditsInput defines the input for the ListUploadAudits use case
type ListUploadAuditsInput struct {
	Page     int
	PageSize int
}

// ListUploadAuditsOutput defines the output for the ListUploadAudits use case
type ListUploadAuditsOutput struct {
	UploadAudits []participantDomain.UploadAudit
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
	uploadAudits, totalCount, err := s.uploadAuditRepository.List(input.Page, input.PageSize)
	if err != nil {
		return ListUploadAuditsOutput{}, fmt.Errorf("failed to list upload audits: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	return ListUploadAuditsOutput{
		UploadAudits: uploadAudits,
		TotalCount:   totalCount,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalPages:   totalPages,
	}, nil
}
