package draw

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GetEligibilityStatsInput represents the input for the GetEligibilityStats use case
type GetEligibilityStatsInput struct {
	DrawDate time.Time
}

// GetEligibilityStatsOutput represents the output from the GetEligibilityStats use case
type GetEligibilityStatsOutput struct {
	TotalParticipants int64
	TotalEntries      int64
	DrawDate          time.Time
	PointsDistribution map[int]int64 // Map of points to count of participants with that many points
}

// GetEligibilityStatsUseCase defines the use case for retrieving eligibility statistics
type GetEligibilityStatsUseCase struct {
	drawRepo       draw.Repository
	participantRepo participant.Repository
}

// NewGetEligibilityStatsUseCase creates a new GetEligibilityStatsUseCase
func NewGetEligibilityStatsUseCase(drawRepo draw.Repository, participantRepo participant.Repository) *GetEligibilityStatsUseCase {
	return &GetEligibilityStatsUseCase{
		drawRepo:       drawRepo,
		participantRepo: participantRepo,
	}
}

// Execute performs the get eligibility statistics use case
func (uc *GetEligibilityStatsUseCase) Execute(ctx context.Context, input GetEligibilityStatsInput) (GetEligibilityStatsOutput, error) {
	// Validate input
	if input.DrawDate.IsZero() {
		return GetEligibilityStatsOutput{}, draw.ErrInvalidDrawDate
	}

	// Get eligible participants for the draw date
	filter := participant.ParticipantFilter{
		DrawDate: input.DrawDate,
	}
	
	// Get total eligible participants
	totalParticipants, err := uc.participantRepo.CountParticipants(ctx, filter)
	if err != nil {
		return GetEligibilityStatsOutput{}, err
	}

	// Get total entries (sum of all points)
	totalEntries, err := uc.participantRepo.CountTotalEntries(ctx, filter)
	if err != nil {
		return GetEligibilityStatsOutput{}, err
	}

	// Get points distribution
	pointsDistribution, err := uc.participantRepo.GetPointsDistribution(ctx, filter)
	if err != nil {
		return GetEligibilityStatsOutput{}, err
	}

	return GetEligibilityStatsOutput{
		TotalParticipants: totalParticipants,
		TotalEntries:      totalEntries,
		DrawDate:          input.DrawDate,
		PointsDistribution: pointsDistribution,
	}, nil
}
