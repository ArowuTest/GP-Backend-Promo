package prize

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	prizeDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// GetPrizeStructureService provides functionality for retrieving prize structures
type GetPrizeStructureService struct {
	prizeRepository prizeDomain.PrizeRepository
}

// NewGetPrizeStructureService creates a new GetPrizeStructureService
func NewGetPrizeStructureService(prizeRepository prizeDomain.PrizeRepository) *GetPrizeStructureService {
	return &GetPrizeStructureService{
		prizeRepository: prizeRepository,
	}
}

// GetPrizeStructureInput defines the input for the GetPrizeStructure use case
type GetPrizeStructureInput struct {
	ID uuid.UUID
}

// GetPrizeStructureOutput defines the output for the GetPrizeStructure use case
type GetPrizeStructureOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	Prizes      []PrizeOutput
	IsActive    bool
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PrizeOutput defines the output for a prize tier
type PrizeOutput struct {
	ID                uuid.UUID
	Name              string
	Description       string
	Value             float64
	Quantity          int
	NumberOfRunnerUps int
}

// GetPrizeStructure retrieves a prize structure by ID
func (s *GetPrizeStructureService) GetPrizeStructure(ctx context.Context, input GetPrizeStructureInput) (*GetPrizeStructureOutput, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("prize structure ID is required")
	}
	
	prizeStructure, err := s.prizeRepository.GetPrizeStructureByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prize structure: %w", err)
	}
	
	// Prepare output
	prizeOutputs := make([]PrizeOutput, 0, len(prizeStructure.Prizes))
	for _, prize := range prizeStructure.Prizes {
		prizeOutputs = append(prizeOutputs, PrizeOutput{
			ID:                prize.ID,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	return &GetPrizeStructureOutput{
		ID:          prizeStructure.ID,
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		StartDate:   prizeStructure.StartDate,
		EndDate:     prizeStructure.EndDate,
		Prizes:      prizeOutputs,
		IsActive:    prizeStructure.IsActive,
		CreatedBy:   prizeStructure.CreatedBy,
		CreatedAt:   prizeStructure.CreatedAt,
		UpdatedAt:   prizeStructure.UpdatedAt,
	}, nil
}
