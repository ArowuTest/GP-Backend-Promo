package draw

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// ListWinnersService provides functionality for retrieving winners
type ListWinnersService struct {
	drawRepository drawDomain.DrawRepository
}

// NewListWinnersService creates a new ListWinnersService
func NewListWinnersService(drawRepository drawDomain.DrawRepository) *ListWinnersService {
	return &ListWinnersService{
		drawRepository: drawRepository,
	}
}

// ListWinnersInput defines the input for the ListWinners use case
type ListWinnersInput struct {
	Page      int
	PageSize  int
	StartDate string
	EndDate   string
}

// ListWinnersOutput defines the output for the ListWinners use case
type ListWinnersOutput struct {
	Winners    []drawDomain.Winner
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ListWinners retrieves winners based on criteria
func (s *ListWinnersService) ListWinners(ctx context.Context, input ListWinnersInput) (ListWinnersOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}

	if input.PageSize < 1 {
		input.PageSize = 10
	}

	// Parse date range if provided
	var startDate, endDate time.Time
	var err error

	if input.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return ListWinnersOutput{}, fmt.Errorf("invalid start date format: %w", err)
		}
	}

	if input.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return ListWinnersOutput{}, fmt.Errorf("invalid end date format: %w", err)
		}
	}

	// Get winners from repository
	winners, totalCount, err := s.drawRepository.ListWinners(ctx, startDate, endDate, input.Page, input.PageSize)
	if err != nil {
		return ListWinnersOutput{}, fmt.Errorf("failed to list winners: %w", err)
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	return ListWinnersOutput{
		Winners:    winners,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
