package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
)

// ParticipantServiceAdapter adapts the participant service to a consistent interface
type ParticipantServiceAdapter struct {
	listParticipantsService   participant.ListParticipantsService
	getParticipantStatsService participant.GetParticipantStatsService
	listUploadAuditsService   participant.ListUploadAuditsService
	uploadParticipantsService participant.UploadParticipantsService
	deleteUploadService       participant.DeleteUploadService
}

// NewParticipantServiceAdapter creates a new ParticipantServiceAdapter
func NewParticipantServiceAdapter(
	listParticipantsService participant.ListParticipantsService,
	getParticipantStatsService participant.GetParticipantStatsService,
	listUploadAuditsService participant.ListUploadAuditsService,
	uploadParticipantsService participant.UploadParticipantsService,
	deleteUploadService participant.DeleteUploadService,
) *ParticipantServiceAdapter {
	return &ParticipantServiceAdapter{
		listParticipantsService:   listParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
		listUploadAuditsService:   listUploadAuditsService,
		uploadParticipantsService: uploadParticipantsService,
		deleteUploadService:       deleteUploadService,
	}
}

// Participant represents a participant
type Participant struct {
	ID        uuid.UUID
	MSISDN    string
	Points    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UploadAudit represents an upload audit
type UploadAudit struct {
	ID             uuid.UUID
	FileName       string
	UploadedBy     uuid.UUID
	UploadedAt     time.Time
	RecordsCount   int
	Status         string
	ErrorMessage   string
	ProcessingTime string
}

// ListParticipantsOutput represents the output of ListParticipants
type ListParticipantsOutput struct {
	Participants []Participant
	Page         int
	PageSize     int
	TotalCount   int
	TotalPages   int
}

// GetParticipantStatsOutput represents the output of GetParticipantStats
type GetParticipantStatsOutput struct {
	TotalParticipants int
	TotalPoints       int
	LastUpdated       time.Time
}

// ListUploadAuditsOutput represents the output of ListUploadAudits
type ListUploadAuditsOutput struct {
	Audits      []UploadAudit
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

// UploadParticipantsOutput represents the output of UploadParticipants
type UploadParticipantsOutput struct {
	ID             uuid.UUID
	FileName       string
	UploadedAt     time.Time
	RecordsCount   int
	Status         string
	ErrorMessage   string
	ProcessingTime string
}

// DeleteUploadOutput represents the output of DeleteUpload
type DeleteUploadOutput struct {
	Success bool
}

// ListParticipants lists participants with pagination and search
func (p *ParticipantServiceAdapter) ListParticipants(ctx context.Context, page, pageSize int, search string) (*ListParticipantsOutput, error) {
	// Call the actual service
	input := participant.ListParticipantsInput{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	output, err := p.listParticipantsService.ListParticipants(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	participants := make([]Participant, 0, len(output.Participants))
	for _, p := range output.Participants {
		participants = append(participants, Participant{
			ID:        p.ID,
			MSISDN:    p.MSISDN,
			Points:    p.Points,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	return &ListParticipantsOutput{
		Participants: participants,
		Page:         output.Page,
		PageSize:     output.PageSize,
		TotalCount:   output.TotalCount,
		TotalPages:   output.TotalPages,
	}, nil
}

// GetParticipantStats gets participant stats
func (p *ParticipantServiceAdapter) GetParticipantStats(ctx context.Context) (*GetParticipantStatsOutput, error) {
	// Call the actual service
	output, err := p.getParticipantStatsService.GetParticipantStats()
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &GetParticipantStatsOutput{
		TotalParticipants: output.TotalParticipants,
		TotalPoints:       output.TotalPoints,
		LastUpdated:       output.LastUpdated,
	}, nil
}

// ListUploadAudits lists upload audits with pagination
func (p *ParticipantServiceAdapter) ListUploadAudits(ctx context.Context, page, pageSize int) (*ListUploadAuditsOutput, error) {
	// Call the actual service
	input := participant.ListUploadAuditsInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := p.listUploadAuditsService.ListUploadAudits(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	audits := make([]UploadAudit, 0, len(output.Audits))
	for _, a := range output.Audits {
		audits = append(audits, UploadAudit{
			ID:             a.ID,
			FileName:       a.FileName,
			UploadedBy:     a.UploadedBy,
			UploadedAt:     a.UploadedAt,
			RecordsCount:   a.RecordsCount,
			Status:         a.Status,
			ErrorMessage:   a.ErrorMessage,
			ProcessingTime: a.ProcessingTime,
		})
	}

	return &ListUploadAuditsOutput{
		Audits:      audits,
		Page:        output.Page,
		PageSize:    output.PageSize,
		TotalCount:  output.TotalCount,
		TotalPages:  output.TotalPages,
	}, nil
}

// UploadParticipants uploads participants
func (p *ParticipantServiceAdapter) UploadParticipants(ctx context.Context, participants []participant.ParticipantInput, uploadedBy uuid.UUID, fileName string) (*UploadParticipantsOutput, error) {
	// Call the actual service
	input := participant.UploadParticipantsInput{
		Participants: participants,
		UploadedBy:   uploadedBy,
		FileName:     fileName,
	}

	output, err := p.uploadParticipantsService.UploadParticipants(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &UploadParticipantsOutput{
		ID:             output.ID,
		FileName:       output.FileName,
		UploadedAt:     output.UploadedAt,
		RecordsCount:   output.RecordsCount,
		Status:         output.Status,
		ErrorMessage:   output.ErrorMessage,
		ProcessingTime: output.ProcessingTime,
	}, nil
}

// DeleteUpload deletes an upload
func (p *ParticipantServiceAdapter) DeleteUpload(ctx context.Context, uploadID uuid.UUID, deletedBy uuid.UUID) (*DeleteUploadOutput, error) {
	// Call the actual service
	input := participant.DeleteUploadInput{
		UploadID:  uploadID,
		DeletedBy: deletedBy,
	}

	output, err := p.deleteUploadService.DeleteUpload(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &DeleteUploadOutput{
		Success: output.Success,
	}, nil
}
