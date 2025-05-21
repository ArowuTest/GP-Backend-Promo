package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// ListWinnersInput represents input for ListWinners
type ListWinnersInput struct {
	Page      int
	PageSize  int
	StartDate string
	EndDate   string
}

// ListWinnersOutput represents output for ListWinners
type ListWinnersOutput struct {
	Winners    []Winner
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

// ListWinnersService handles listing winners
type ListWinnersService struct {
	repository Repository
}

// NewListWinnersService creates a new ListWinnersService
func NewListWinnersService(repository Repository) *ListWinnersService {
	return &ListWinnersService{
		repository: repository,
	}
}

// ListWinners lists winners with pagination
func (s *ListWinnersService) ListWinners(ctx context.Context, input ListWinnersInput) (ListWinnersOutput, error) {
	// For now, return mock data
	mockWinners := []Winner{
		{
			ID:            uuid.New(),
			DrawID:        uuid.New(),
			MSISDN:        "234*****789",
			PrizeTierID:   uuid.New(),
			PrizeTierName: "First Prize",
			PrizeValue:    1000000,
			Status:        "Active",
			PaymentStatus: "Pending",
			PaymentNotes:  "",
			PaidAt:        "",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	
	return ListWinnersOutput{
		Winners:    mockWinners,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalCount: len(mockWinners),
		TotalPages: 1,
	}, nil
}
