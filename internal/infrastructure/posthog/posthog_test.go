package posthog_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/posthog"
)

// PostHogClientInterface defines the interface for the PostHog client
type PostHogClientInterface interface {
	CreateCohort(ctx context.Context, name string, filters []posthog.Filter) (string, error)
	GetCohort(ctx context.Context, id string) (posthog.Cohort, error)
	ListCohorts(ctx context.Context) ([]posthog.Cohort, error)
	GetCohortPersons(ctx context.Context, cohortID string) ([]posthog.Person, error)
	CaptureEvent(ctx context.Context, distinctID string, event string, properties map[string]interface{}) error
}

// MockPostHogClient is a mock implementation of the PostHog client
type MockPostHogClient struct {
	mock.Mock
}

func (m *MockPostHogClient) CreateCohort(ctx context.Context, name string, filters []posthog.Filter) (string, error) {
	args := m.Called(ctx, name, filters)
	return args.String(0), args.Error(1)
}

func (m *MockPostHogClient) GetCohort(ctx context.Context, id string) (posthog.Cohort, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(posthog.Cohort), args.Error(1)
}

func (m *MockPostHogClient) ListCohorts(ctx context.Context) ([]posthog.Cohort, error) {
	args := m.Called(ctx)
	return args.Get(0).([]posthog.Cohort), args.Error(1)
}

func (m *MockPostHogClient) GetCohortPersons(ctx context.Context, cohortID string) ([]posthog.Person, error) {
	args := m.Called(ctx, cohortID)
	return args.Get(0).([]posthog.Person), args.Error(1)
}

func (m *MockPostHogClient) CaptureEvent(ctx context.Context, distinctID string, event string, properties map[string]interface{}) error {
	args := m.Called(ctx, distinctID, event, properties)
	return args.Error(0)
}

// TestPostHogDataSource_GetEligibleParticipants tests the GetEligibleParticipants method
func TestPostHogDataSource_GetEligibleParticipants(t *testing.T) {
	// Create mock client
	mockClient := new(MockPostHogClient)
	
	// Create test date
	testDate := time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)
	testDateStr := testDate.Format("2006-01-02")
	cohortName := "Eligible Participants - " + testDateStr
	cohortID := uuid.New().String()
	
	// Setup mock expectations
	mockClient.On("ListCohorts", mock.Anything).Return([]posthog.Cohort{}, nil)
	mockClient.On("CreateCohort", mock.Anything, cohortName, mock.Anything).Return(cohortID, nil)
	
	// Setup test persons
	testPersons := []posthog.Person{
		{
			ID:         uuid.New().String(),
			DistinctID: "2347012345678",
			Properties: map[string]interface{}{
				"points":          float64(5),
				"recharge_amount": float64(500),
			},
			CreatedAt: time.Now(),
		},
		{
			ID:         uuid.New().String(),
			DistinctID: "2347087654321",
			Properties: map[string]interface{}{
				"points":          float64(10),
				"recharge_amount": float64(1000),
			},
			CreatedAt: time.Now(),
		},
	}
	
	mockClient.On("GetCohortPersons", mock.Anything, cohortID).Return(testPersons, nil)
	
	// Create cohort generator and data source
	cohortGenerator := posthog.NewCohortGenerator(mockClient)
	dataSource := posthog.NewPostHogDataSource(mockClient, cohortGenerator)
	
	// Call the method being tested
	participants, err := dataSource.GetEligibleParticipants(context.Background(), testDate)
	
	// Assert expectations
	assert.NoError(t, err)
	assert.Len(t, participants, 2)
	assert.Equal(t, "2347012345678", participants[0].MSISDN)
	assert.Equal(t, 5, participants[0].Points)
	assert.Equal(t, 500.0, participants[0].RechargeAmount)
	assert.Equal(t, "2347087654321", participants[1].MSISDN)
	assert.Equal(t, 10, participants[1].Points)
	assert.Equal(t, 1000.0, participants[1].RechargeAmount)
	
	mockClient.AssertExpectations(t)
}

// TestPostHogDataSource_RecordDrawResult tests the RecordDrawResult method
func TestPostHogDataSource_RecordDrawResult(t *testing.T) {
	// Create mock client
	mockClient := new(MockPostHogClient)
	
	// Create test winners
	drawID := uuid.New().String()
	testWinners := []draw.Winner{
		{
			ID:            uuid.New(),
			DrawID:        uuid.New(),
			MSISDN:        "2347012345678",
			PrizeTierID:   uuid.New(),
			PrizeTierName: "First Prize",
			PrizeValue:    100000.0,
			Status:        "PENDING",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			DrawID:        uuid.New(),
			MSISDN:        "2347087654321",
			PrizeTierID:   uuid.New(),
			PrizeTierName: "Second Prize",
			PrizeValue:    50000.0,
			Status:        "PENDING",
			IsRunnerUp:    true,
			RunnerUpRank:  1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	
	// Setup mock expectations
	mockClient.On("CaptureEvent", mock.Anything, "2347012345678", "draw_winner", mock.Anything).Return(nil)
	mockClient.On("CaptureEvent", mock.Anything, "2347087654321", "draw_runner_up", mock.Anything).Return(nil)
	
	// Create cohort generator and data source
	cohortGenerator := posthog.NewCohortGenerator(mockClient)
	dataSource := posthog.NewPostHogDataSource(mockClient, cohortGenerator)
	
	// Call the method being tested
	err := dataSource.RecordDrawResult(context.Background(), drawID, testWinners)
	
	// Assert expectations
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestCohortGenerator_EnsureDailyCohort tests the EnsureDailyCohort method
func TestCohortGenerator_EnsureDailyCohort(t *testing.T) {
	// Create mock client
	mockClient := new(MockPostHogClient)
	
	// Create test date
	testDate := time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)
	testDateStr := testDate.Format("2006-01-02")
	cohortName := "Eligible Participants - " + testDateStr
	cohortID := uuid.New().String()
	
	// Test case 1: Cohort doesn't exist yet
	mockClient.On("ListCohorts", mock.Anything).Return([]posthog.Cohort{}, nil).Once()
	mockClient.On("CreateCohort", mock.Anything, cohortName, mock.Anything).Return(cohortID, nil).Once()
	
	// Create cohort generator
	cohortGenerator := posthog.NewCohortGenerator(mockClient)
	
	// Call the method being tested
	resultID, err := cohortGenerator.EnsureDailyCohort(context.Background(), testDate)
	
	// Assert expectations
	assert.NoError(t, err)
	assert.Equal(t, cohortID, resultID)
	
	// Test case 2: Cohort already exists
	existingCohort := posthog.Cohort{
		ID:        cohortID,
		Name:      cohortName,
		CreatedAt: time.Now(),
		Count:     100,
	}
	
	mockClient.On("ListCohorts", mock.Anything).Return([]posthog.Cohort{existingCohort}, nil).Once()
	
	// Call the method being tested again
	resultID, err = cohortGenerator.EnsureDailyCohort(context.Background(), testDate)
	
	// Assert expectations
	assert.NoError(t, err)
	assert.Equal(t, cohortID, resultID)
	
	mockClient.AssertExpectations(t)
}
