package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// PrizeServiceAdapter adapts the prize service to a consistent interface
type PrizeServiceAdapter struct {
	createPrizeStructureService *prize.CreatePrizeStructureService
	getPrizeStructureService    *prize.GetPrizeStructureService
	listPrizeStructuresService  *prize.ListPrizeStructuresService
	updatePrizeStructureService *prize.UpdatePrizeStructureService
	deletePrizeStructureService *prize.DeletePrizeStructureService
}

// NewPrizeServiceAdapter creates a new PrizeServiceAdapter
func NewPrizeServiceAdapter(
	createPrizeStructureService *prize.CreatePrizeStructureService,
	getPrizeStructureService *prize.GetPrizeStructureService,
	listPrizeStructuresService *prize.ListPrizeStructuresService,
	updatePrizeStructureService *prize.UpdatePrizeStructureService,
	deletePrizeStructureService *prize.DeletePrizeStructureService,
) *PrizeServiceAdapter {
	return &PrizeServiceAdapter{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		updatePrizeStructureService: updatePrizeStructureService,
		deletePrizeStructureService: deletePrizeStructureService,
	}
}

// CreatePrizeStructure creates a new prize structure
func (p *PrizeServiceAdapter) CreatePrizeStructure(
	ctx context.Context,
	name string,
	description string,
	startDate time.Time,
	endDate time.Time,
	prizes []entity.PrizeInput,
	createdBy uuid.UUID,
	isActive bool,
) (*entity.PrizeStructure, error) {
	// Convert prizes to domain model
	// Use entity.PrizeInput directly instead of undefined prize.CreatePrizeInput
	domainPrizes := make([]entity.PrizeInput, 0, len(prizes))
	for _, p := range prizes {
		domainPrizes = append(domainPrizes, entity.PrizeInput{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create input for the service
	input := prize.CreatePrizeStructureInput{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Prizes:      []prize.PrizeInput{},  // Use empty slice of correct type
		CreatedBy:   createdBy,
		IsActive:    isActive,
	}
	
	// Convert domain prizes to application layer prizes
	for _, p := range domainPrizes {
		input.Prizes = append(input.Prizes, prize.PrizeInput{
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create prize structure
	output, err := p.createPrizeStructureService.CreatePrizeStructure(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert prizes to entity model
	entityPrizes := make([]entity.Prize, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		entityPrizes = append(entityPrizes, entity.Prize{
			ID:                p.ID,
			PrizeStructureID:  output.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	result := &entity.PrizeStructure{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      entityPrizes,
		IsActive:    output.IsActive,
		CreatedBy:   output.CreatedBy,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}

	return result, nil
}

// GetPrizeStructure gets a prize structure by ID
func (p *PrizeServiceAdapter) GetPrizeStructure(
	ctx context.Context,
	id uuid.UUID,
) (*entity.PrizeStructure, error) {
	// Create input for the service
	input := prize.GetPrizeStructureInput{
		ID: id,
	}

	// Get prize structure
	output, err := p.getPrizeStructureService.GetPrizeStructure(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert prizes to entity model
	entityPrizes := make([]entity.Prize, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		entityPrizes = append(entityPrizes, entity.Prize{
			ID:                p.ID,
			PrizeStructureID:  output.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	result := &entity.PrizeStructure{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      entityPrizes,
		IsActive:    output.IsActive,
		CreatedBy:   output.CreatedBy,
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
	input := prize.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}

	// Get prize structures
	output, err := p.listPrizeStructuresService.ListPrizeStructures(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert prize structures to entity model
	prizeStructures := make([]entity.PrizeStructure, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		// Convert prizes to entity model
		entityPrizes := make([]entity.Prize, 0, len(ps.Prizes))
		for _, p := range ps.Prizes {
			entityPrizes = append(entityPrizes, entity.Prize{
				ID:                p.ID,
				PrizeStructureID:  ps.ID,
				Name:              p.Name,
				Description:       p.Description,
				Value:             p.Value,
				Quantity:          p.Quantity,
				NumberOfRunnerUps: p.NumberOfRunnerUps,
			})
		}

		prizeStructures = append(prizeStructures, entity.PrizeStructure{
			ID:          ps.ID,
			Name:        ps.Name,
			Description: ps.Description,
			StartDate:   ps.StartDate,
			EndDate:     ps.EndDate,
			Prizes:      entityPrizes,
			IsActive:    ps.IsActive,
			CreatedBy:   ps.CreatedBy,
			CreatedAt:   ps.CreatedAt,
			UpdatedAt:   ps.UpdatedAt,
		})
	}

	// Create response
	result := &entity.PaginatedPrizeStructures{
		PrizeStructures: prizeStructures,
		Page:            output.Page,
		PageSize:        output.PageSize,
		TotalCount:      output.TotalCount,
		TotalPages:      output.TotalPages,
	}

	return result, nil
}

// UpdatePrizeStructure updates a prize structure
func (p *PrizeServiceAdapter) UpdatePrizeStructure(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description string,
	startDate time.Time,
	endDate time.Time,
	prizes []entity.PrizeInput,
	updatedBy uuid.UUID,
	isActive bool,
) (*entity.PrizeStructure, error) {
	// Convert prizes to domain model
	domainPrizes := make([]prize.UpdatePrizeInput, 0, len(prizes))
	for _, p := range prizes {
		domainPrizes = append(domainPrizes, prize.UpdatePrizeInput{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create input for the service
	input := prize.UpdatePrizeStructureInput{
		ID:          id,
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Prizes:      domainPrizes,
		UpdatedBy:   updatedBy,
		IsActive:    isActive,
	}

	// Update prize structure
	output, err := p.updatePrizeStructureService.UpdatePrizeStructure(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert prizes to entity model
	entityPrizes := make([]entity.Prize, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		entityPrizes = append(entityPrizes, entity.Prize{
			ID:                p.ID,
			PrizeStructureID:  output.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	result := &entity.PrizeStructure{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      entityPrizes,
		IsActive:    output.IsActive,
		UpdatedBy:   output.UpdatedBy,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}

	return result, nil
}

// DeletePrizeStructure deletes a prize structure
func (p *PrizeServiceAdapter) DeletePrizeStructure(
	ctx context.Context,
	id uuid.UUID,
	deletedBy uuid.UUID,
) error {
	// Create input for the service using domain type
	input := struct {
		ID        uuid.UUID
		DeletedBy uuid.UUID
	}{
		ID:        id,
		DeletedBy: deletedBy,
	}

	// Delete prize structure
	return p.deletePrizeStructureService.DeletePrizeStructure(ctx, input)
}
