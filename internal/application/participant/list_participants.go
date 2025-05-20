package participant

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// ListParticipantsService provides functionality for retrieving participants
type ListParticipantsService struct {
	participantRepository participantDomain.ParticipantRepository
}

// NewListParticipantsService creates a new ListParticipantsService
func NewListParticipantsService(participantRepository participantDomain.ParticipantRepository) *ListParticipantsService {
	return &ListParticipantsService{
		participantRepository: participantRepository,
	}
}

// ListParticipantsInput defines the input for the ListParticipants use case
type ListParticipantsInput struct {
	Page     int
	PageSize int
}

// ListParticipantsOutput defines the output for the ListParticipants use case
type ListParticipantsOutput struct {
	Participants []participantDomain.Participant
	TotalCount   int
	Page         int
	PageSize     int
	TotalPages   int
}

// ListParticipants retrieves participants based on criteria
func (s *ListParticipantsService) ListParticipants(ctx context.Context, input ListParticipantsInput) (ListParticipantsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}

	if input.PageSize < 1 {
		input.PageSize = 10
	}

	// Get participants from repository
	participants, totalCount, err := s.participantRepository.List(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListParticipantsOutput{}, fmt.Errorf("failed to list participants: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	return ListParticipantsOutput{
		Participants: participants,
		TotalCount:   totalCount,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalPages:   totalPages,
	}, nil
}
