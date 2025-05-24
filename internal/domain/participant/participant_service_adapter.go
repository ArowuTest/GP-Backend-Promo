package participant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// ParticipantServiceAdapter adapts the participant service to a consistent interface
type ParticipantServiceAdapter struct {
	// Internal services
	uploadParticipantsService *UploadParticipantsService
	getParticipantStatsService *GetParticipantStatsService
	listParticipantsService    *ListParticipantsService
	listUploadAuditsService    *ListUploadAuditsService
	deleteUploadService        *DeleteUploadService
}

// NewParticipantServiceAdapter creates a new ParticipantServiceAdapter
func NewParticipantServiceAdapter(
	uploadParticipantsService *UploadParticipantsService,
	getParticipantStatsService *GetParticipantStatsService,
	listParticipantsService *ListParticipantsService,
	listUploadAuditsService *ListUploadAuditsService,
	deleteUploadService *DeleteUploadService,
) *ParticipantServiceAdapter {
	return &ParticipantServiceAdapter{
		uploadParticipantsService: uploadParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
		listParticipantsService:    listParticipantsService,
		listUploadAuditsService:    listUploadAuditsService,
		deleteUploadService:        deleteUploadService,
	}
}

// UploadParticipants uploads participants from a file
func (p *ParticipantServiceAdapter) UploadParticipants(
	ctx context.Context,
	fileContent []byte,
	filename string,
	uploadedBy uuid.UUID,
) (*entity.UploadAudit, error) {
	// Create input for the service
	input := UploadParticipantsInput{
		FileContent: fileContent,
		Filename:    filename,
		UploadedBy:  uploadedBy,
	}

	// Upload participants
	output, err := p.uploadParticipantsService.UploadParticipants(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.UploadAudit{
		ID:           output.ID,
		FileName:     output.Filename,
		RecordCount:  output.RecordCount,
		UploadedBy:   output.UploadedBy,
		UploadDate:   output.UploadDate,
		Status:       "Completed",
		TotalRows:    output.RecordCount,
		CreatedAt:    output.UploadDate,
		UpdatedAt:    output.UploadDate,
	}

	return result, nil
}

// GetParticipantStats gets statistics about participants
func (p *ParticipantServiceAdapter) GetParticipantStats(
	ctx context.Context,
) (*entity.ParticipantStats, error) {
	// Create input for the service
	input := GetParticipantStatsInput{}

	// Get participant stats
	output, err := p.getParticipantStatsService.GetParticipantStats(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.ParticipantStats{
		TotalParticipants: output.TotalParticipants,
		TotalPoints:       output.TotalPoints,
		LastUploadDate:    output.LastUploadDate,
		LastUpdated:       output.LastUpdated,
	}

	return result, nil
}

// ListParticipants gets a list of participants with pagination
func (p *ParticipantServiceAdapter) ListParticipants(
	ctx context.Context,
	page, pageSize int,
	phoneNumber string,
) (*entity.PaginatedParticipants, error) {
	// Create input for the service
	input := ListParticipantsInput{
		Page:        page,
		PageSize:    pageSize,
		PhoneNumber: phoneNumber,
	}

	// Get participants
	output, err := p.listParticipantsService.ListParticipants(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.PaginatedParticipants{
		Participants: output.Participants,
		Page:         output.Page,
		PageSize:     output.PageSize,
		TotalCount:   output.TotalCount,
		TotalPages:   output.TotalPages,
	}

	return result, nil
}

// ListUploadAudits gets a list of upload audits with pagination
func (p *ParticipantServiceAdapter) ListUploadAudits(
	ctx context.Context,
	page, pageSize int,
) (*entity.PaginatedUploadAudits, error) {
	// Create input for the service
	input := ListUploadAuditsInput{
		Page:     page,
		PageSize: pageSize,
	}

	// Get upload audits
	output, err := p.listUploadAuditsService.ListUploadAudits(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.PaginatedUploadAudits{
		UploadAudits: output.UploadAudits,
		Page:         output.Page,
		PageSize:     output.PageSize,
		TotalCount:   output.TotalCount,
		TotalPages:   output.TotalPages,
	}

	return result, nil
}

// DeleteUpload deletes an upload and its associated participants
func (p *ParticipantServiceAdapter) DeleteUpload(
	ctx context.Context,
	id uuid.UUID,
) error {
	// Create input for the service
	input := DeleteUploadInput{
		ID: id,
	}

	// Delete upload
	return p.deleteUploadService.DeleteUpload(ctx, input)
}
