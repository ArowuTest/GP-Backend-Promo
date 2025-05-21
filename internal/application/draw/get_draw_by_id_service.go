package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Winner represents a draw winner
type Winner struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	PrizeTierName string
	PrizeValue    float64
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        string
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Draw represents a draw
type Draw struct {
	ID                   uuid.UUID
	DrawDate             time.Time
	PrizeStructureID     uuid.UUID
	Status               string
	TotalEligibleMSISDNs int
	TotalEntries         int
	ExecutedBy           uuid.UUID
	Winners              []Winner
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GetDrawByIDInput represents input for GetDrawByID
type GetDrawByIDInput struct {
	ID uuid.UUID
}

// GetDrawByIDOutput represents output for GetDrawByID
type GetDrawByIDOutput struct {
	ID                   uuid.UUID
	DrawDate             time.Time
	PrizeStructureID     uuid.UUID
	Status               string
	TotalEligibleMSISDNs int
	TotalEntries         int
	ExecutedBy           uuid.UUID
	Winners              []Winner
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GetDrawByIDService handles retrieving a draw by ID
type GetDrawByIDService struct {
	repository Repository
}

// NewGetDrawByIDService creates a new GetDrawByIDService
func NewGetDrawByIDService(repository Repository) *GetDrawByIDService {
	return &GetDrawByIDService{
		repository: repository,
	}
}

// GetDrawByID retrieves a draw by ID
func (s *GetDrawByIDService) GetDrawByID(ctx context.Context, input GetDrawByIDInput) (GetDrawByIDOutput, error) {
	// For now, return mock data
	mockWinners := []Winner{
		{
			ID:            uuid.New(),
			DrawID:        input.ID,
			MSISDN:        "234*****789",
			PrizeTierID:   uuid.New(),
			PrizeTierName: "First Prize",
			PrizeValue:    1000000,
			Status:        "Active",
			PaymentStatus: "Pending",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	
	return GetDrawByIDOutput{
		ID:                   input.ID,
		DrawDate:             time.Now(),
		PrizeStructureID:     uuid.New(),
		Status:               "Completed",
		TotalEligibleMSISDNs: 1000,
		TotalEntries:         5000,
		ExecutedBy:           uuid.New(),
		Winners:              mockWinners,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}
