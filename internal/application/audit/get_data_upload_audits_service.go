package audit

import (
	"context"
	"time"
	
	auditDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GetDataUploadAuditsInput represents input for GetDataUploadAudits
type GetDataUploadAuditsInput struct {
	Page      int
	PageSize  int
	StartDate time.Time
	EndDate   time.Time
}

// GetDataUploadAuditsOutput represents output for GetDataUploadAudits
type GetDataUploadAuditsOutput struct {
	Audits     []auditDomain.DataUploadAudit
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

// GetDataUploadAuditsService handles retrieving data upload audit logs
type GetDataUploadAuditsService struct {
	repository Repository
}

// NewGetDataUploadAuditsService creates a new GetDataUploadAuditsService
func NewGetDataUploadAuditsService(repository Repository) *GetDataUploadAuditsService {
	return &GetDataUploadAuditsService{
		repository: repository,
	}
}

// GetDataUploadAudits retrieves data upload audit logs
func (s *GetDataUploadAuditsService) GetDataUploadAudits(ctx context.Context, input GetDataUploadAuditsInput) (GetDataUploadAuditsOutput, error) {
	// Implementation using domain types
	audits, total, err := s.repository.GetDataUploadAudits(ctx, input.Page, input.PageSize, input.StartDate, input.EndDate)
	if err != nil {
		return GetDataUploadAuditsOutput{}, err
	}
	
	// Convert to output format
	auditOutputs := make([]auditDomain.DataUploadAudit, len(audits))
	for i, audit := range audits {
		auditOutputs[i] = *audit
	}
	
	totalPages := total / input.PageSize
	if total%input.PageSize > 0 {
		totalPages++
	}
	
	return GetDataUploadAuditsOutput{
		Audits:     auditOutputs,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}, nil
}
