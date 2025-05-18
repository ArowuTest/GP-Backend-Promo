package participant

import (
	"context"
	"fmt"
	"time"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GetParticipantStatsService provides functionality for retrieving participant statistics
type GetParticipantStatsService struct {
	participantRepository participant.ParticipantRepository
}

// NewGetParticipantStatsService creates a new GetParticipantStatsService
func NewGetParticipantStatsService(
	participantRepository participant.ParticipantRepository,
) *GetParticipantStatsService {
	return &GetParticipantStatsService{
		participantRepository: participantRepository,
	}
}

// GetParticipantStatsInput defines the input for the GetParticipantStats use case
type GetParticipantStatsInput struct {
	StartDate string // Format: YYYY-MM-DD
	EndDate   string // Format: YYYY-MM-DD
}

// GetParticipantStatsOutput defines the output for the GetParticipantStats use case
type GetParticipantStatsOutput struct {
	TotalParticipants int     `json:"totalParticipants"`
	TotalPoints       int     `json:"totalPoints"`
	AveragePoints     float64 `json:"averagePoints"`
	StartDate         string  `json:"startDate"`
	EndDate           string  `json:"endDate"`
}

// GetParticipantStats retrieves statistics for participants
func (s *GetParticipantStatsService) GetParticipantStats(ctx context.Context, input GetParticipantStatsInput) (*GetParticipantStatsOutput, error) {
	if input.StartDate == "" {
		return nil, fmt.Errorf("start date is required")
	}
	
	if input.EndDate == "" {
		return nil, fmt.Errorf("end date is required")
	}
	
	// Parse dates
	startDate, err := parseDate(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	
	// Parse end date but we don't use it currently as the repository method doesn't need it
	_, err = parseDate(input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	
	// Get participant stats
	totalParticipants, totalPoints, _, err := s.participantRepository.GetStats(startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant stats: %w", err)
	}
	
	// Calculate average points
	var averagePoints float64
	if totalParticipants > 0 {
		averagePoints = float64(totalPoints) / float64(totalParticipants)
	}
	
	return &GetParticipantStatsOutput{
		TotalParticipants: totalParticipants,
		TotalPoints:       totalPoints,
		AveragePoints:     averagePoints,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
	}, nil
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
