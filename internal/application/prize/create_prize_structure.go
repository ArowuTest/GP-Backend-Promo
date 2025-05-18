package application

import (
	"time"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// CreatePrizeStructureUseCase represents the use case for creating a prize structure
type CreatePrizeStructureUseCase struct {
	prizeRepository prize.PrizeRepository
}

// NewCreatePrizeStructureUseCase creates a new CreatePrizeStructureUseCase
func NewCreatePrizeStructureUseCase(
	prizeRepository prize.PrizeRepository,
) *CreatePrizeStructureUseCase {
	return &CreatePrizeStructureUseCase{
		prizeRepository: prizeRepository,
	}
}

// CreatePrizeStructureInput represents the input for the create prize structure use case
type CreatePrizeStructureInput struct {
	Name        string
	Description string
	IsActive    bool
	ValidFrom   time.Time
	ValidTo     *time.Time
	Prizes      []PrizeTierInput
}

// PrizeTierInput represents the input for a prize tier
type PrizeTierInput struct {
	Rank        int
	Name        string
	Description string
	Value       string
	ValueNGN    float64
	Quantity    int
}

// CreatePrizeStructureOutput represents the output of the create prize structure use case
type CreatePrizeStructureOutput struct {
	PrizeStructure *prize.PrizeStructure
}

// Execute creates a new prize structure
func (uc *CreatePrizeStructureUseCase) Execute(input CreatePrizeStructureInput) (*CreatePrizeStructureOutput, error) {
	// Create prize structure entity
	prizeStructureID := uuid.New()
	prizeStructure := &prize.PrizeStructure{
		ID:          prizeStructureID,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    input.IsActive,
		ValidFrom:   input.ValidFrom,
		ValidTo:     input.ValidTo,
		Prizes:      make([]prize.PrizeTier, 0, len(input.Prizes)),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create prize tiers
	for _, tierInput := range input.Prizes {
		tier := prize.PrizeTier{
			ID:               uuid.New(),
			PrizeStructureID: prizeStructureID,
			Rank:             tierInput.Rank,
			Name:             tierInput.Name,
			Description:      tierInput.Description,
			Value:            tierInput.Value,
			ValueNGN:         tierInput.ValueNGN,
			Quantity:         tierInput.Quantity,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// Validate prize tier
		if err := prize.ValidatePrizeTier(&tier); err != nil {
			return nil, prize.NewPrizeError(prize.ErrInvalidPrizeTier, "Invalid prize tier", err)
		}

		prizeStructure.Prizes = append(prizeStructure.Prizes, tier)
	}

	// Validate prize structure
	if err := prize.ValidatePrizeStructure(prizeStructure); err != nil {
		return nil, prize.NewPrizeError(prize.ErrInvalidPrizeStructure, "Invalid prize structure", err)
	}

	// Save prize structure
	if err := uc.prizeRepository.CreatePrizeStructure(prizeStructure); err != nil {
		return nil, prize.NewPrizeError("PRIZE_STRUCTURE_CREATION_FAILED", "Failed to create prize structure", err)
	}

	return &CreatePrizeStructureOutput{
		PrizeStructure: prizeStructure,
	}, nil
}
