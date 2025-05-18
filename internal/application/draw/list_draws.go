package draw

import (
	"context"
	"fmt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// ListDrawsService provides functionality for listing draws
type ListDrawsService struct {
	drawRepository draw.DrawRepository
}

// NewListDrawsService creates a new ListDrawsService
func NewListDrawsService(drawRepository draw.DrawRepository) *ListDrawsService {
	return &ListDrawsService{
		drawRepository: drawRepository,
	}
}

// ListDrawsInput defines the input for the ListDraws use case
type ListDrawsInput struct {
	Page     int
	PageSize int
}

// ListDrawsOutput defines the output for the ListDraws use case
type ListDrawsOutput struct {
	Draws      []draw.Draw
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ListDraws retrieves a paginated list of draws
func (s *ListDrawsService) ListDraws(ctx context.Context, input ListDrawsInput) (*ListDrawsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	
	draws, totalCount, err := s.drawRepository.List(input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list draws: %w", err)
	}
	
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}
	
	return &ListDrawsOutput{
		Draws:      draws,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
