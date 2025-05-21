package draw

import (
	"context"
	"time"
	
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
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
	Winners    []drawDomain.Winner
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
	// Implementation using domain types
	winners, total, err := s.repository.ListWinners(ctx, input.Page, input.PageSize, input.StartDate, input.EndDate)
	if err != nil {
		return ListWinnersOutput{}, err
	}
	
	// Convert to output format
	winnerOutputs := make([]drawDomain.Winner, len(winners))
	for i, winner := range winners {
		winnerOutputs[i] = *winner
	}
	
	totalPages := total / input.PageSize
	if total%input.PageSize > 0 {
		totalPages++
	}
	
	return ListWinnersOutput{
		Winners:    winnerOutputs,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}, nil
}
