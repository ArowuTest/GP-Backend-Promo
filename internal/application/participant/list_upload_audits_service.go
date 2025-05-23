package participant

import (
	"context"
	
	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// ListUploadAuditsInput represents input for ListUploadAudits
type ListUploadAuditsInput struct {
	Page     int
	PageSize int
}

// ListUploadAuditsOutput represents output for ListUploadAudits
type ListUploadAuditsOutput struct {
	Audits     []participantDomain.UploadAudit
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

// ListUploadAuditsService handles listing upload audits
type ListUploadAuditsService struct {
	repository Repository
}

// NewListUploadAuditsService creates a new ListUploadAuditsService
func NewListUploadAuditsService(repository Repository) *ListUploadAuditsService {
	return &ListUploadAuditsService{
		repository: repository,
	}
}

// ListUploadAudits lists upload audits with pagination
func (s *ListUploadAuditsService) ListUploadAudits(ctx context.Context, input ListUploadAuditsInput) (ListUploadAuditsOutput, error) {
	// Implementation using domain types
	audits, total, err := s.repository.ListUploadAudits(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListUploadAuditsOutput{}, err
	}
	
	// Convert to output format
	auditOutputs := make([]participantDomain.UploadAudit, len(audits))
	for i, audit := range audits {
		auditOutputs[i] = *audit
	}
	
	totalPages := total / input.PageSize
	if total%input.PageSize > 0 {
		totalPages++
	}
	
	return ListUploadAuditsOutput{
		Audits:     auditOutputs,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}, nil
}
