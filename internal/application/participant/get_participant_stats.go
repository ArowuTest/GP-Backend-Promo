package participant

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GetParticipantStatsInput represents the input for the GetParticipantStats use case
type GetParticipantStatsInput struct {
	StartDate time.Time
	EndDate   time.Time
}

// GetParticipantStatsOutput represents the output from the GetParticipantStats use case
type GetParticipantStatsOutput struct {
	TotalParticipants int64
	TotalPoints       int64
	AveragePoints     float64
	DateStats         []DateStat
}

// DateStat represents statistics for a specific date
type DateStat struct {
	Date              time.Time
	ParticipantCount  int64
	TotalPoints       int64
	AveragePointsPerParticipant float64
}

// GetParticipantStatsUseCase defines the use case for retrieving participant statistics
type GetParticipantStatsUseCase struct {
	participantRepo participant.Repository
}

// NewGetParticipantStatsUseCase creates a new GetParticipantStatsUseCase
func NewGetParticipantStatsUseCase(participantRepo participant.Repository) *GetParticipantStatsUseCase {
	return &GetParticipantStatsUseCase{
		participantRepo: participantRepo,
	}
}

// Execute performs the get participant statistics use case
func (uc *GetParticipantStatsUseCase) Execute(ctx context.Context, input GetParticipantStatsInput) (GetParticipantStatsOutput, error) {
	// Validate input
	if input.StartDate.IsZero() || input.EndDate.IsZero() {
		return GetParticipantStatsOutput{}, participant.ErrInvalidDateRange
	}

	// Get total participants in date range
	totalParticipants, err := uc.participantRepo.CountParticipantsInDateRange(ctx, input.StartDate, input.EndDate)
	if err != nil {
		return GetParticipantStatsOutput{}, err
	}

	// Get total points in date range
	totalPoints, err := uc.participantRepo.CountTotalPointsInDateRange(ctx, input.StartDate, input.EndDate)
	if err != nil {
		return GetParticipantStatsOutput{}, err
	}

	// Calculate average points per participant
	var averagePoints float64
	if totalParticipants > 0 {
		averagePoints = float64(totalPoints) / float64(totalParticipants)
	}

	// Get daily statistics
	dateStats, err := uc.participantRepo.GetDailyStats(ctx, input.StartDate, input.EndDate)
	if err != nil {
		return GetParticipantStatsOutput{}, err
	}

	// Convert repository data to use case output format
	stats := make([]DateStat, len(dateStats))
	for i, stat := range dateStats {
		var avgPoints float64
		if stat.ParticipantCount > 0 {
			avgPoints = float64(stat.TotalPoints) / float64(stat.ParticipantCount)
		}
		
		stats[i] = DateStat{
			Date:              stat.Date,
			ParticipantCount:  stat.ParticipantCount,
			TotalPoints:       stat.TotalPoints,
			AveragePointsPerParticipant: avgPoints,
		}
	}

	return GetParticipantStatsOutput{
		TotalParticipants: totalParticipants,
		TotalPoints:       totalPoints,
		AveragePoints:     averagePoints,
		DateStats:         stats,
	}, nil
}
