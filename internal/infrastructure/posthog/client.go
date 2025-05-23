package posthog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client represents a PostHog API client
type Client struct {
	apiKey     string
	projectID  string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new PostHog client
func NewClient(apiKey, projectID, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		projectID:  projectID,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Filter represents a PostHog cohort filter
type Filter struct {
	Property string      `json:"property"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Type     string      `json:"type,omitempty"`
}

// Cohort represents a PostHog cohort
type Cohort struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Filters     []Filter  `json:"filters"`
	CreatedAt   time.Time `json:"created_at"`
	Count       int       `json:"count"`
}

// Person represents a PostHog person (participant)
type Person struct {
	ID         string                 `json:"id"`
	DistinctID string                 `json:"distinct_id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
}

// CreateCohort creates a new cohort in PostHog
func (c *Client) CreateCohort(ctx context.Context, name string, filters []Filter) (string, error) {
	url := fmt.Sprintf("%s/api/projects/%s/cohorts/", c.baseURL, c.projectID)
	
	payload := map[string]interface{}{
		"name":        name,
		"filters":     filters,
		"is_static":   false,
		"description": fmt.Sprintf("Auto-generated cohort for %s", name),
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cohort payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create cohort, status code: %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	cohortID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid cohort ID in response")
	}
	
	return cohortID, nil
}

// GetCohort retrieves a cohort by ID
func (c *Client) GetCohort(ctx context.Context, id string) (Cohort, error) {
	url := fmt.Sprintf("%s/api/projects/%s/cohorts/%s/", c.baseURL, c.projectID, id)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Cohort{}, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Cohort{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return Cohort{}, fmt.Errorf("failed to get cohort, status code: %d", resp.StatusCode)
	}
	
	var cohort Cohort
	if err := json.NewDecoder(resp.Body).Decode(&cohort); err != nil {
		return Cohort{}, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return cohort, nil
}

// ListCohorts lists all cohorts
func (c *Client) ListCohorts(ctx context.Context) ([]Cohort, error) {
	url := fmt.Sprintf("%s/api/projects/%s/cohorts/", c.baseURL, c.projectID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list cohorts, status code: %d", resp.StatusCode)
	}
	
	var result struct {
		Results []Cohort `json:"results"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Results, nil
}

// GetCohortPersons retrieves all persons in a cohort
func (c *Client) GetCohortPersons(ctx context.Context, cohortID string) ([]Person, error) {
	url := fmt.Sprintf("%s/api/projects/%s/cohorts/%s/persons/", c.baseURL, c.projectID, cohortID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get cohort persons, status code: %d", resp.StatusCode)
	}
	
	var result struct {
		Results []Person `json:"results"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Results, nil
}

// CaptureEvent sends an event to PostHog
func (c *Client) CaptureEvent(ctx context.Context, distinctID string, event string, properties map[string]interface{}) error {
	url := fmt.Sprintf("%s/capture/", c.baseURL)
	
	if properties == nil {
		properties = make(map[string]interface{})
	}
	
	// Add project API key to properties
	properties["api_key"] = c.apiKey
	
	payload := map[string]interface{}{
		"event":      event,
		"distinct_id": distinctID,
		"properties": properties,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to capture event, status code: %d", resp.StatusCode)
	}
	
	return nil
}
