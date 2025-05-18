package prize

import (
	"context"
	"errors"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// GetPrizeStructureInput represents the input for the GetPrizeStructure use case
type GetPrizeStructureInput struct {
	PrizeStructureID string
}

// GetPrizeStructureOutput represents the output from the GetPrizeStructure use case
type GetPrizeStructureOutput struct {
	PrizeStructure prize.PrizeStructure
}

// GetPrizeStructureUseCase defines the use case for retrieving a prize structure
type GetPrizeStructureUseCase struct {
	prizeRepo prize.Repository
}

// NewGetPrizeStructureUseCase creates a new GetPrizeStructureUseCase
func NewGetPrizeStructureUseCase(prizeRepo prize.Repository) *GetPrizeStructureUseCase {
	return &GetPrizeStructureUseCase{
		prizeRepo: prizeRepo,
	}
}

// Execute performs the get prize structure use case
func (uc *GetPrizeStructureUseCase) Execute(ctx context.Context, input GetPrizeStructureInput) (GetPrizeStructureOutput, error) {
	// Validate input
	if input.PrizeStructureID == "" {
		return GetPrizeStructureOutput{}, errors.New("prize structure ID is required")
	}

	// Get prize structure from repository
	prizeStructure, err := uc.prizeRepo.GetPrizeStructureByID(ctx, input.PrizeStructureID)
	if err != nil {
		return GetPrizeStructureOutput{}, err
	}

	return GetPrizeStructureOutput{
		PrizeStructure: prizeStructure,
	}, nil
}
