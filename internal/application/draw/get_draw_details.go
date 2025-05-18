package draw

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GetDrawDetailsInput represents the input for the GetDrawDetails use case
type GetDrawDetailsInput struct {
	DrawID string
}

// GetDrawDetailsOutput represents the output from the GetDrawDetails use case
type GetDrawDetailsOutput struct {
	Draw       draw.Draw
	Winners    []draw.Winner
	RunnerUps  []draw.RunnerUp
	TotalPrize float64
}

// GetDrawDetailsUseCase defines the use case for retrieving draw details
type GetDrawDetailsUseCase struct {
	drawRepo draw.Repository
}

// NewGetDrawDetailsUseCase creates a new GetDrawDetailsUseCase
func NewGetDrawDetailsUseCase(drawRepo draw.Repository) *GetDrawDetailsUseCase {
	return &GetDrawDetailsUseCase{
		drawRepo: drawRepo,
	}
}

// Execute performs the get draw details use case
func (uc *GetDrawDetailsUseCase) Execute(ctx context.Context, input GetDrawDetailsInput) (GetDrawDetailsOutput, error) {
	// Validate input
	if input.DrawID == "" {
		return GetDrawDetailsOutput{}, draw.ErrInvalidDrawID
	}

	// Get draw from repository
	d, err := uc.drawRepo.GetDrawByID(ctx, input.DrawID)
	if err != nil {
		return GetDrawDetailsOutput{}, err
	}

	// Get winners for this draw
	winners, err := uc.drawRepo.GetWinnersByDrawID(ctx, input.DrawID)
	if err != nil {
		return GetDrawDetailsOutput{}, err
	}

	// Get runner-ups for this draw
	runnerUps, err := uc.drawRepo.GetRunnerUpsByDrawID(ctx, input.DrawID)
	if err != nil {
		return GetDrawDetailsOutput{}, err
	}

	// Calculate total prize amount
	var totalPrize float64
	for _, winner := range winners {
		totalPrize += winner.PrizeAmount
	}

	return GetDrawDetailsOutput{
		Draw:       d,
		Winners:    winners,
		RunnerUps:  runnerUps,
		TotalPrize: totalPrize,
	}, nil
}
