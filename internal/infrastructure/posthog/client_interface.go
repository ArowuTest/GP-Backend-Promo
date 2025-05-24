package posthog

import (
	"context"
)

// PostHogClientInterface defines the interface for the PostHog client
type PostHogClientInterface interface {
	CreateCohort(ctx context.Context, name string, filters []Filter) (string, error)
	GetCohort(ctx context.Context, id string) (Cohort, error)
	ListCohorts(ctx context.Context) ([]Cohort, error)
	GetCohortPersons(ctx context.Context, cohortID string) ([]Person, error)
	CaptureEvent(ctx context.Context, distinctID string, event string, properties map[string]interface{}) error
}

// Ensure Client implements PostHogClientInterface
var _ PostHogClientInterface = (*Client)(nil)
