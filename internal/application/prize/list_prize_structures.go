package prize

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// ListPrizeStructuresInput represents the input for the ListPrizeStructures use case
type ListPrizeStructuresInput struct {
	Active   *bool
	Page     int
	PageSize int
}

// ListPrizeStructuresOutput represents the output from the ListPrizeStructures use case
type ListPrizeStructuresOutput struct {
	PrizeStructures []prize.PrizeStructure
	Total           int64
	Page            int
	PageSize        int
}

// ListPrizeStructuresUseCase defines the use case for listing prize structures
type ListPrizeStructuresUseCase struct {
	prizeRepo prize.Repository
}

// NewListPrizeStructuresUseCase creates a new ListPrizeStructuresUseCase
func NewListPrizeStructuresUseCase(prizeRepo prize.Repository) *ListPrizeStructuresUseCase {
	return &ListPrizeStructuresUseCase{
		prizeRepo: prizeRepo,
	}
}

// Execute performs the list prize structures use case
func (uc *ListPrizeStructuresUseCase) Execute(ctx context.Context, input ListPrizeStructuresInput) (ListPrizeStructuresOutput, error) {
	// Set default page size if not provided
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	// Set default page if not provided
	if input.Page <= 0 {
		input.Page = 1
	}

	// Prepare filter criteria
	filter := prize.PrizeStructureFilter{
		Active:   input.Active,
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	// Get prize structures from repository
	prizeStructures, err := uc.prizeRepo.ListPrizeStructures(ctx, filter)
	if err != nil {
		return ListPrizeStructuresOutput{}, err
	}

	// Get total count for pagination
	total, err := uc.prizeRepo.CountPrizeStructures(ctx, filter)
	if err != nil {
		return ListPrizeStructuresOutput{}, err
	}

	return ListPrizeStructuresOutput{
		PrizeStructures: prizeStructures,
		Total:           total,
		Page:            input.Page,
		PageSize:        input.PageSize,
	}, nil
}
