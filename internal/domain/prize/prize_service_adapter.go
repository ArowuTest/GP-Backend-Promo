package prize

import (
	"context"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// PrizeServiceAdapter adapts the prize service to a consistent interface
type PrizeServiceAdapter struct {
	// Internal services
	createPrizeStructureService *CreatePrizeStructureService
	getPrizeStructureService    *GetPrizeStructureService
	listPrizeStructuresService  *ListPrizeStructuresService
	deletePrizeStructureService DeletePrizeStructureService
}

// NewPrizeServiceAdapter creates a new PrizeServiceAdapter
func NewPrizeServiceAdapter(
	createPrizeStructureService *CreatePrizeStructureService,
	getPrizeStructureService *GetPrizeStructureService,
	listPrizeStructuresService *ListPrizeStructuresService,
	deletePrizeStructureService DeletePrizeStructureService,
) *PrizeServiceAdapter {
	return &PrizeServiceAdapter{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		deletePrizeStructureService: deletePrizeStructureService,
	}
}

// CreatePrizeStructure creates a new prize structure
func (p *PrizeServiceAdapter) CreatePrizeStructure(
	ctx context.Context,
	name string,
	description string,
	prizes []entity.PrizeInput,
	createdBy uuid.UUID,
) (*entity.PrizeStructure, error) {
	// Convert entity.PrizeInput to CreatePrizeInput
	createPrizes := make([]CreatePrizeInput, 0, len(prizes))
	for _, p := range prizes {
		createPrizes = append(createPrizes, CreatePrizeInput{
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}
	
	// Create input for the service
	input := CreatePrizeStructureInput{
		Name:        name,
		Description: description,
		Prizes:      createPrizes,
		CreatedBy:   createdBy,
	}

	// Create prize structure
	output, err := p.createPrizeStructureService.CreatePrizeStructure(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert output prizes to entity prizes
	entityPrizes := make([]entity.Prize, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		entityPrizes = append(entityPrizes, entity.Prize{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Value:       p.Value,
			Quantity:    p.Quantity,
			Position:    p.Position,
		})
	}

	// Create response
	result := &entity.PrizeStructure{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Prizes:      entityPrizes,
		CreatedBy:   output.CreatedBy,
		CreatedAt:   output.CreatedAt,
	}

	return result, nil
}

// GetPrizeStructure gets a prize structure by ID
func (p *PrizeServiceAdapter) GetPrizeStructure(
	ctx context.Context,
	id uuid.UUID,
) (*entity.PrizeStructure, error) {
	// Create input for the service
	input := GetPrizeStructureInput{
		ID: id,
	}

	// Get prize structure
	output, err := p.getPrizeStructureService.GetPrizeStructure(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert output prizes to entity prizes
	entityPrizes := make([]entity.Prize, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		entityPrizes = append(entityPrizes, entity.Prize{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Value:       p.Value,
			Quantity:    p.Quantity,
			Position:    p.Position,
		})
	}

	// Create response
	result := &entity.PrizeStructure{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Prizes:      entityPrizes,
		CreatedBy:   output.CreatedBy,
		UpdatedBy:   output.UpdatedBy,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}

	return result, nil
}

// ListPrizeStructures gets a list of prize structures with pagination
func (p *PrizeServiceAdapter) ListPrizeStructures(
	ctx context.Context,
	page, pageSize int,
) (*entity.PaginatedPrizeStructures, error) {
	// Create input for the service
	input := ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}

	// Get prize structures
	output, err := p.listPrizeStructuresService.ListPrizeStructures(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.PaginatedPrizeStructures{
		PrizeStructures: output.PrizeStructures,
		Page:            output.Page,
		PageSize:        output.PageSize,
		TotalCount:      output.TotalCount,
		TotalPages:      output.TotalPages,
	}

	return result, nil
}

// DeletePrizeStructure deletes a prize structure
func (p *PrizeServiceAdapter) DeletePrizeStructure(
	ctx context.Context,
	id uuid.UUID,
) error {
	// Create input for the service
	input := DeletePrizeStructureInput{
		ID: id,
	}

	// Delete prize structure using the implementation function
	return DeletePrizeStructureImpl(ctx, input)
}
