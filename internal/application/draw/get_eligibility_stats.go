package draw

import (
	"context"
	"fmt"
	"time"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GetEligibilityStatsService provides functionality for retrieving eligibility statistics
type GetEligibilityStatsService struct {
	drawRepository draw.DrawRepository
	participantRepository participant.ParticipantRepository
}

// NewGetEligibilityStatsService creates a new GetEligibilityStatsService
func NewGetEligibilityStatsService(
	drawRepository draw.DrawRepository,
	participantRepository participant.ParticipantRepository,
) *GetEligibilityStatsService {
	return &GetEligibilityStatsService{
		drawRepository: drawRepository,
		participantRepository: participantRepository,
	}
}

// GetEligibilityStatsInput defines the input for the GetEligibilityStats use case
type GetEligibilityStatsInput struct {
	Date string // Format: YYYY-MM-DD
}

// GetEligibilityStatsOutput defines the output for the GetEligibilityStats use case
type GetEligibilityStatsOutput struct {
	TotalEligibleMSISDNs int
	TotalEntries int
	Date string
}

// GetEligibilityStats retrieves eligibility statistics for a specific date
func (s *GetEligibilityStatsService) GetEligibilityStats(ctx context.Context, input GetEligibilityStatsInput) (*GetEligibilityStatsOutput, error) {
	if input.Date == "" {
		return nil, fmt.Errorf("date is required")
	}
	
	// Parse date
	date, err := parseDate(input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	
	// Get eligibility stats
	totalEligibleMSISDNs, totalEntries, err := s.drawRepository.GetEligibilityStats(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligibility stats: %w", err)
	}
	
	return &GetEligibilityStatsOutput{
		TotalEligibleMSISDNs: totalEligibleMSISDNs,
		TotalEntries: totalEntries,
		Date: input.Date,
	}, nil
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
