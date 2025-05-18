package prize

import (
	"context"
	"errors"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// UpdatePrizeStructureInput represents the input for the UpdatePrizeStructure use case
type UpdatePrizeStructureInput struct {
	PrizeStructureID string
	Name             string
	Description      string
	Active           bool
	Prizes           []PrizeInput
	UpdatedBy        string
}

// PrizeInput represents the input for a prize in the prize structure
type PrizeInput struct {
	Name        string
	Description string
	Amount      float64
	Quantity    int
	Rank        int
}

// UpdatePrizeStructureOutput represents the output from the UpdatePrizeStructure use case
type UpdatePrizeStructureOutput struct {
	PrizeStructure prize.PrizeStructure
}

// UpdatePrizeStructureUseCase defines the use case for updating a prize structure
type UpdatePrizeStructureUseCase struct {
	prizeRepo prize.Repository
}

// NewUpdatePrizeStructureUseCase creates a new UpdatePrizeStructureUseCase
func NewUpdatePrizeStructureUseCase(prizeRepo prize.Repository) *UpdatePrizeStructureUseCase {
	return &UpdatePrizeStructureUseCase{
		prizeRepo: prizeRepo,
	}
}

// Execute performs the update prize structure use case
func (uc *UpdatePrizeStructureUseCase) Execute(ctx context.Context, input UpdatePrizeStructureInput) (UpdatePrizeStructureOutput, error) {
	// Validate input
	if input.PrizeStructureID == "" {
		return UpdatePrizeStructureOutput{}, errors.New("prize structure ID is required")
	}
	if input.Name == "" {
		return UpdatePrizeStructureOutput{}, errors.New("name is required")
	}
	if len(input.Prizes) == 0 {
		return UpdatePrizeStructureOutput{}, errors.New("at least one prize is required")
	}
	if input.UpdatedBy == "" {
		return UpdatePrizeStructureOutput{}, errors.New("updater information is required")
	}

	// Get existing prize structure
	existingPrizeStructure, err := uc.prizeRepo.GetPrizeStructureByID(ctx, input.PrizeStructureID)
	if err != nil {
		return UpdatePrizeStructureOutput{}, err
	}

	// Update prize structure fields
	existingPrizeStructure.Name = input.Name
	existingPrizeStructure.Description = input.Description
	existingPrizeStructure.Active = input.Active
	existingPrizeStructure.UpdatedBy = input.UpdatedBy

	// Convert prize inputs to domain prizes
	prizes := make([]prize.Prize, len(input.Prizes))
	for i, p := range input.Prizes {
		if p.Name == "" {
			return UpdatePrizeStructureOutput{}, errors.New("prize name is required")
		}
		if p.Amount <= 0 {
			return UpdatePrizeStructureOutput{}, errors.New("prize amount must be greater than zero")
		}
		if p.Quantity <= 0 {
			return UpdatePrizeStructureOutput{}, errors.New("prize quantity must be greater than zero")
		}

		prizes[i] = prize.Prize{
			Name:        p.Name,
			Description: p.Description,
			Amount:      p.Amount,
			Quantity:    p.Quantity,
			Rank:        p.Rank,
		}
	}

	// Update prizes in prize structure
	existingPrizeStructure.Prizes = prizes

	// Update prize structure in repository
	updatedPrizeStructure, err := uc.prizeRepo.UpdatePrizeStructure(ctx, existingPrizeStructure)
	if err != nil {
		return UpdatePrizeStructureOutput{}, err
	}

	return UpdatePrizeStructureOutput{
		PrizeStructure: updatedPrizeStructure,
	}, nil
}
