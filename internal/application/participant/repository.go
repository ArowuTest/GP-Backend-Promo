package participant

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Repository defines the interface for participant repository
type Repository interface {
	ListParticipants(ctx context.Context, page, pageSize int) ([]Participant, int, error)
	ListUploadAudits(ctx context.Context, page, pageSize int) ([]UploadAudit, int, error)
	DeleteUpload(ctx context.Context, uploadID, deletedBy uuid.UUID) (bool, error)
	UploadParticipants(ctx context.Context, participants []ParticipantInput, uploadedBy uuid.UUID, fileName string) (uuid.UUID, int, int, int, error)
	GetParticipantStats(ctx context.Context, startDate, endDate string) (int, int, error)
}

// ParticipantInput represents input for uploading a participant
type ParticipantInput struct {
	MSISDN         string
	RechargeAmount float64
	RechargeDate   string
}

// UploadParticipantsInput represents input for UploadParticipants
type UploadParticipantsInput struct {
	Participants []ParticipantInput
	UploadedBy   uuid.UUID
	FileName     string
}

// UploadParticipantsOutput represents output for UploadParticipants
type UploadParticipantsOutput struct {
	AuditID           uuid.UUID
	TotalRowsProcessed int
	SuccessfulRows    int
	ErrorCount        int
	DuplicatesSkipped int
	ErrorDetails      string
	Status            string
	UploadedAt        time.Time
}

// UploadParticipantsService handles uploading participants
type UploadParticipantsService struct {
	repository Repository
}

// NewUploadParticipantsService creates a new UploadParticipantsService
func NewUploadParticipantsService(repository Repository) *UploadParticipantsService {
	return &UploadParticipantsService{
		repository: repository,
	}
}

// UploadParticipants uploads participants
func (s *UploadParticipantsService) UploadParticipants(ctx context.Context, input UploadParticipantsInput) (UploadParticipantsOutput, error) {
	// For now, return mock data
	return UploadParticipantsOutput{
		AuditID:           uuid.New(),
		TotalRowsProcessed: len(input.Participants),
		SuccessfulRows:    len(input.Participants),
		ErrorCount:        0,
		DuplicatesSkipped: 0,
		ErrorDetails:      "",
		Status:            "Completed",
		UploadedAt:        time.Now(),
	}, nil
}

// GetParticipantStatsInput represents input for GetParticipantStats
type GetParticipantStatsInput struct {
	StartDate string
	EndDate   string
}

// GetParticipantStatsOutput represents output for GetParticipantStats
type GetParticipantStatsOutput struct {
	StartDate         string
	EndDate           string
	TotalParticipants int
	TotalPoints       int
}

// GetParticipantStatsService handles getting participant statistics
type GetParticipantStatsService struct {
	repository Repository
}

// NewGetParticipantStatsService creates a new GetParticipantStatsService
func NewGetParticipantStatsService(repository Repository) *GetParticipantStatsService {
	return &GetParticipantStatsService{
		repository: repository,
	}
}

// GetParticipantStats gets participant statistics
func (s *GetParticipantStatsService) GetParticipantStats(ctx context.Context, input GetParticipantStatsInput) (GetParticipantStatsOutput, error) {
	// For now, return mock data
	return GetParticipantStatsOutput{
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		TotalParticipants: 1000,
		TotalPoints:       5000,
	}, nil
}
