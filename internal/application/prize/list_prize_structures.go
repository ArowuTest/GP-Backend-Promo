package prize

import (
	"context"
	"fmt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// ListPrizeStructuresService provides functionality for listing prize structures
type ListPrizeStructuresService struct {
	prizeRepository prize.PrizeRepository
}

// NewListPrizeStructuresService creates a new ListPrizeStructuresService
func NewListPrizeStructuresService(prizeRepository prize.PrizeRepository) *ListPrizeStructuresService {
	return &ListPrizeStructuresService{
		prizeRepository: prizeRepository,
	}
}

// ListPrizeStructuresInput defines the input for the ListPrizeStructures use case
type ListPrizeStructuresInput struct {
	Page     int
	PageSize int
}

// ListPrizeStructuresOutput defines the output for the ListPrizeStructures use case
type ListPrizeStructuresOutput struct {
	PrizeStructures []prize.PrizeStructure
	TotalCount      int
	Page            int
	PageSize        int
	TotalPages      int
}

// ListPrizeStructures retrieves a paginated list of prize structures
func (s *ListPrizeStructuresService) ListPrizeStructures(ctx context.Context, input ListPrizeStructuresInput) (*ListPrizeStructuresOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	
	prizeStructures, totalCount, err := s.prizeRepository.ListPrizeStructures(input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list prize structures: %w", err)
	}
	
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}
	
	return &ListPrizeStructuresOutput{
		PrizeStructures: prizeStructures,
		TotalCount:      totalCount,
		Page:            input.Page,
		PageSize:        input.PageSize,
		TotalPages:      totalPages,
	}, nil
}
