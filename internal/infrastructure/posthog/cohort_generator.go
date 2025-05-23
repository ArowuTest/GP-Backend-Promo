package posthog

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
)

// CohortGenerator handles the creation and management of daily cohorts
type CohortGenerator struct {
	client *Client
}

// NewCohortGenerator creates a new CohortGenerator
func NewCohortGenerator(client *Client) *CohortGenerator {
	return &CohortGenerator{
		client: client,
	}
}

// EnsureDailyCohort ensures a cohort exists for the specified date
// If it doesn't exist, it creates one
func (cg *CohortGenerator) EnsureDailyCohort(ctx context.Context, date time.Time) (string, error) {
	// Format date for cohort name
	dateStr := date.Format("2006-01-02")
	cohortName := fmt.Sprintf("Eligible Participants - %s", dateStr)
	
	// Check if cohort already exists
	cohorts, err := cg.client.ListCohorts(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list cohorts: %w", err)
	}
	
	// Look for existing cohort with this name
	for _, cohort := range cohorts {
		if cohort.Name == cohortName {
			return cohort.ID, nil
		}
	}
	
	// Create filters for the cohort
	// These filters define eligibility criteria for the draw
	filters := []Filter{
		{
			Property: "recharge_date",
			Operator: "is_date_before",
			Value:    dateStr,
			Type:     "person",
		},
		{
			Property: "points",
			Operator: "gt",
			Value:    0,
			Type:     "person",
		},
		{
			Property: "blacklisted",
			Operator: "is_not",
			Value:    true,
			Type:     "person",
		},
	}
	
	// Create the cohort
	cohortID, err := cg.client.CreateCohort(ctx, cohortName, filters)
	if err != nil {
		return "", fmt.Errorf("failed to create cohort: %w", err)
	}
	
	return cohortID, nil
}

// GetCohortStats retrieves statistics for a cohort
func (cg *CohortGenerator) GetCohortStats(ctx context.Context, cohortID string) (int, int, error) {
	// Get cohort details
	cohort, err := cg.client.GetCohort(ctx, cohortID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get cohort: %w", err)
	}
	
	// Get persons in the cohort
	persons, err := cg.client.GetCohortPersons(ctx, cohortID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get cohort persons: %w", err)
	}
	
	// Calculate total entries based on points
	totalEntries := 0
	for _, person := range persons {
		points := 0
		if pointsVal, ok := person.Properties["points"]; ok {
			if p, ok := pointsVal.(float64); ok {
				points = int(p)
			}
		}
		totalEntries += points
	}
	
	return len(persons), totalEntries, nil
}

// CreateTestCohort creates a test cohort with sample data
func (cg *CohortGenerator) CreateTestCohort(ctx context.Context) (string, error) {
	// Create a test cohort
	cohortName := fmt.Sprintf("Test Cohort - %s", uuid.New().String())
	
	// Simple filter for test cohort
	filters := []Filter{
		{
			Property: "is_test",
			Operator: "exact",
			Value:    true,
			Type:     "person",
		},
	}
	
	// Create the cohort
	cohortID, err := cg.client.CreateCohort(ctx, cohortName, filters)
	if err != nil {
		return "", fmt.Errorf("failed to create test cohort: %w", err)
	}
	
	// Add test persons to the cohort
	// In a real implementation, this would be done through the PostHog API
	// For testing purposes, we'll assume the persons are added separately
	
	return cohortID, nil
}
