package posthog

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// PostHogDataSource implements the DrawDataSource interface using PostHog as the data source
type PostHogDataSource struct {
	client          PostHogClientInterface
	cohortGenerator *CohortGenerator
}

// NewPostHogDataSource creates a new PostHogDataSource
func NewPostHogDataSource(client PostHogClientInterface, cohortGenerator *CohortGenerator) *PostHogDataSource {
	return &PostHogDataSource{
		client:          client,
		cohortGenerator: cohortGenerator,
	}
}

// GetEligibleParticipants retrieves eligible participants for a specific date from PostHog
func (ds *PostHogDataSource) GetEligibleParticipants(ctx context.Context, date time.Time) ([]participant.Participant, error) {
	// Ensure we have a cohort for this date
	cohortID, err := ds.cohortGenerator.EnsureDailyCohort(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure daily cohort: %w", err)
	}
	
	// Get persons from the cohort
	persons, err := ds.client.GetCohortPersons(ctx, cohortID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cohort persons: %w", err)
	}
	
	// Convert PostHog persons to domain participants
	participants := make([]participant.Participant, 0, len(persons))
	for _, person := range persons {
		// Extract MSISDN from distinct_id
		msisdn := person.DistinctID
		
		// Extract points from properties
		points := 0
		if pointsVal, ok := person.Properties["points"]; ok {
			if p, ok := pointsVal.(float64); ok {
				points = int(p)
			}
		}
		
		// Extract recharge amount from properties
		rechargeAmount := 0.0
		if amountVal, ok := person.Properties["recharge_amount"]; ok {
			if a, ok := amountVal.(float64); ok {
				rechargeAmount = a
			}
		}
		
		// Create participant entity
		participant := participant.Participant{
			ID:             uuid.New(),
			MSISDN:         msisdn,
			Points:         points,
			RechargeAmount: rechargeAmount,
			RechargeDate:   date,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		
		participants = append(participants, participant)
	}
	
	return participants, nil
}

// GetParticipantEntries calculates the number of entries a participant has based on points
func (ds *PostHogDataSource) GetParticipantEntries(ctx context.Context, msisdn string) (int, error) {
	// This would typically query PostHog for the participant's points
	// For now, we'll implement a simplified version that assumes points are stored in PostHog
	
	// In a real implementation, we would query PostHog for the participant's properties
	// and extract the points value
	
	// For demonstration purposes, we'll return a placeholder value
	// In production, this would be replaced with actual PostHog queries
	return 5, nil
}

// RecordDrawResult records draw results in PostHog for analytics
func (ds *PostHogDataSource) RecordDrawResult(ctx context.Context, drawID string, winners []draw.Winner) error {
	// Record each winner as an event in PostHog
	for _, winner := range winners {
		properties := map[string]interface{}{
			"draw_id":         drawID,
			"prize_tier_id":   winner.PrizeTierID.String(),
			"prize_tier_name": winner.PrizeTierName,
			"prize_value":     winner.PrizeValue,
			"is_runner_up":    winner.IsRunnerUp,
			"runner_up_rank":  winner.RunnerUpRank,
		}
		
		eventName := "draw_winner"
		if winner.IsRunnerUp {
			eventName = "draw_runner_up"
		}
		
		if err := ds.client.CaptureEvent(ctx, winner.MSISDN, eventName, properties); err != nil {
			return fmt.Errorf("failed to capture winner event: %w", err)
		}
	}
	
	return nil
}
