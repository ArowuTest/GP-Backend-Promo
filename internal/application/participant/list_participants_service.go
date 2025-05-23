package participant

import (
	"context"
	
	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// ListParticipantsInput represents input for ListParticipants
type ListParticipantsInput struct {
	Page     int
	PageSize int
}

// ListParticipantsOutput represents output for ListParticipants
type ListParticipantsOutput struct {
	Participants []participantDomain.Participant
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
	// Implementation using domain types
	participants, total, err := s.repository.ListParticipants(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListParticipantsOutput{}, err
	}
	
	// Convert to output format
	participantOutputs := make([]participantDomain.Participant, len(participants))
	for i, participant := range participants {
		participantOutputs[i] = *participant
	}
	
	totalPages := total / input.PageSize
	if total%input.PageSize > 0 {
		totalPages++
	}
	
	return ListParticipantsOutput{
		Participants: participantOutputs,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalCount:   total,
		TotalPages:   totalPages,
	}, nil
}
