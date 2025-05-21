package participant

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Participant represents a participant record
type Participant struct {
	ID             uuid.UUID
	MSISDN         string
	Points         int
	RechargeAmount float64
	RechargeDate   time.Time
	CreatedAt      time.Time
	UploadID       uuid.UUID
}

// ListParticipantsInput represents input for ListParticipants
type ListParticipantsInput struct {
	Page     int
	PageSize int
}

// ListParticipantsOutput represents output for ListParticipants
type ListParticipantsOutput struct {
	Participants []Participant
	Page         int
	PageSize     int
	TotalCount   int
	TotalPages   int
}

// ListParticipantsService handles listing participants
type ListParticipantsService struct {
	repository Repository
}

// NewListParticipantsService creates a new ListParticipantsService
func NewListParticipantsService(repository Repository) *ListParticipantsService {
	return &ListParticipantsService{
		repository: repository,
	}
}

// ListParticipants lists participants with pagination
func (s *ListParticipantsService) ListParticipants(ctx context.Context, input ListParticipantsInput) (ListParticipantsOutput, error) {
	// For now, return mock data
	mockParticipants := []Participant{
		{
			ID:             uuid.New(),
			MSISDN:         "234*****789",
			Points:         5,
			RechargeAmount: 500,
			RechargeDate:   time.Now(),
			CreatedAt:      time.Now(),
			UploadID:       uuid.New(),
		},
	}
	
	return ListParticipantsOutput{
		Participants: mockParticipants,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalCount:   len(mockParticipants),
		TotalPages:   1,
	}, nil
}
