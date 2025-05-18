package draw

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// ListDrawsInput represents the input for the ListDraws use case
type ListDrawsInput struct {
	StartDate time.Time
	EndDate   time.Time
	Status    string
	Page      int
	PageSize  int
}

// ListDrawsOutput represents the output from the ListDraws use case
type ListDrawsOutput struct {
	Draws    []draw.Draw
	Total    int64
	Page     int
	PageSize int
}

// ListDrawsUseCase defines the use case for listing draws
type ListDrawsUseCase struct {
	drawRepo draw.Repository
}

// NewListDrawsUseCase creates a new ListDrawsUseCase
func NewListDrawsUseCase(drawRepo draw.Repository) *ListDrawsUseCase {
	return &ListDrawsUseCase{
		drawRepo: drawRepo,
	}
}

// Execute performs the list draws use case
func (uc *ListDrawsUseCase) Execute(ctx context.Context, input ListDrawsInput) (ListDrawsOutput, error) {
	// Set default page size if not provided
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	// Set default page if not provided
	if input.Page <= 0 {
		input.Page = 1
	}

	// Prepare filter criteria
	filter := draw.DrawFilter{
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Status:    input.Status,
		Page:      input.Page,
		PageSize:  input.PageSize,
	}

	// Get draws from repository
	draws, err := uc.drawRepo.ListDraws(ctx, filter)
	if err != nil {
		return ListDrawsOutput{}, err
	}

	// Get total count for pagination
	total, err := uc.drawRepo.CountDraws(ctx, filter)
	if err != nil {
		return ListDrawsOutput{}, err
	}

	return ListDrawsOutput{
		Draws:    draws,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}
