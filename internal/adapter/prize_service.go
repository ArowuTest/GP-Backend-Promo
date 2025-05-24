package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
)

// PrizeServiceAdapter adapts the prize service to a consistent interface
type PrizeServiceAdapter struct {
	createPrizeStructureService prize.CreatePrizeStructureService
	getPrizeStructureService    prize.GetPrizeStructureService
	listPrizeStructuresService  prize.ListPrizeStructuresService
	updatePrizeStructureService prize.UpdatePrizeStructureService
	deletePrizeStructureService prize.DeletePrizeStructureService
}

// NewPrizeServiceAdapter creates a new PrizeServiceAdapter
func NewPrizeServiceAdapter(
	createPrizeStructureService prize.CreatePrizeStructureService,
	getPrizeStructureService prize.GetPrizeStructureService,
	listPrizeStructuresService prize.ListPrizeStructuresService,
	updatePrizeStructureService prize.UpdatePrizeStructureService,
	deletePrizeStructureService prize.DeletePrizeStructureService,
) *PrizeServiceAdapter {
	return &PrizeServiceAdapter{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		updatePrizeStructureService: updatePrizeStructureService,
		deletePrizeStructureService: deletePrizeStructureService,
	}
}

// PrizeItem represents a prize
type PrizeItem struct {
	ID                string
	Name              string
	Description       string
	Value             float64
	Quantity          int
	NumberOfRunnerUps int
}

// PrizeStructure represents a prize structure
type PrizeStructure struct {
	ID          string
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Prizes      []PrizeItem
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreatePrizeStructureOutput represents the output of CreatePrizeStructure
type CreatePrizeStructureOutput struct {
	ID          string
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Prizes      []PrizeItem
	IsActive    bool
}

// GetPrizeStructureOutput represents the output of GetPrizeStructure
type GetPrizeStructureOutput struct {
	ID          string
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	Prizes      []PrizeItem
	IsActive    bool
}

// ListPrizeStructuresOutput represents the output of ListPrizeStructures
type ListPrizeStructuresOutput struct {
	PrizeStructures []PrizeStructure
	Page            int
	PageSize        int
	TotalCount      int
	TotalPages      int
}

// UpdatePrizeStructureOutput represents the output of UpdatePrizeStructure
type UpdatePrizeStructureOutput struct {
	ID          string
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Prizes      []PrizeItem
	IsActive    bool
}

// CreatePrizeStructure creates a prize structure
func (p *PrizeServiceAdapter) CreatePrizeStructure(
	ctx context.Context,
	name string,
	description string,
	validFrom string,
	validTo string,
	prizes []prize.CreatePrizeInput,
	createdBy uuid.UUID,
	isActive bool,
) (*CreatePrizeStructureOutput, error) {
	// Call the actual service
	input := prize.CreatePrizeStructureInput{
		Name:        name,
		Description: description,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Prizes:      prizes,
		CreatedBy:   createdBy,
		IsActive:    isActive,
	}

	output, err := p.createPrizeStructureService.CreatePrizeStructure(input)
	if err != nil {
		return nil, err
	}

	// Convert prizes for response
	prizesOutput := make([]PrizeItem, 0, len(output.Prizes))
	for _, prize := range output.Prizes {
		prizesOutput = append(prizesOutput, PrizeItem{
			ID:                prize.ID,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}

	// Return response
	return &CreatePrizeStructureOutput{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      prizesOutput,
		IsActive:    output.IsActive,
	}, nil
}

// GetPrizeStructure gets a prize structure by ID
func (p *PrizeServiceAdapter) GetPrizeStructure(ctx context.Context, id uuid.UUID) (*GetPrizeStructureOutput, error) {
	// Call the actual service
	output, err := p.getPrizeStructureService.GetPrizeStructure(id)
	if err != nil {
		return nil, err
	}

	// Convert prizes for response
	prizesOutput := make([]PrizeItem, 0, len(output.Prizes))
	for _, prize := range output.Prizes {
		prizesOutput = append(prizesOutput, PrizeItem{
			ID:                prize.ID,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}

	// Return response
	return &GetPrizeStructureOutput{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      prizesOutput,
		IsActive:    output.IsActive,
	}, nil
}

// ListPrizeStructures lists prize structures with pagination
func (p *PrizeServiceAdapter) ListPrizeStructures(ctx context.Context, page, pageSize int) (*ListPrizeStructuresOutput, error) {
	// Call the actual service
	input := prize.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := p.listPrizeStructuresService.ListPrizeStructures(input)
	if err != nil {
		return nil, err
	}

	// Convert prize structures for response
	prizeStructuresOutput := make([]PrizeStructure, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		// Convert prizes for response
		prizesOutput := make([]PrizeItem, 0, len(ps.Prizes))
		for _, prize := range ps.Prizes {
			prizesOutput = append(prizesOutput, PrizeItem{
				ID:                prize.ID,
				Name:              prize.Name,
				Description:       prize.Description,
				Value:             prize.Value,
				Quantity:          prize.Quantity,
				NumberOfRunnerUps: prize.NumberOfRunnerUps,
			})
		}

		prizeStructuresOutput = append(prizeStructuresOutput, PrizeStructure{
			ID:          ps.ID,
			Name:        ps.Name,
			Description: ps.Description,
			StartDate:   ps.StartDate.Format("2006-01-02"),
			EndDate:     ps.EndDate.Format("2006-01-02"),
			Prizes:      prizesOutput,
			IsActive:    ps.IsActive,
			CreatedAt:   ps.CreatedAt,
			UpdatedAt:   ps.UpdatedAt,
		})
	}

	// Return response
	return &ListPrizeStructuresOutput{
		PrizeStructures: prizeStructuresOutput,
		Page:            output.Page,
		PageSize:        output.PageSize,
		TotalCount:      output.TotalCount,
		TotalPages:      output.TotalPages,
	}, nil
}

// UpdatePrizeStructure updates a prize structure
func (p *PrizeServiceAdapter) UpdatePrizeStructure(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description string,
	validFrom string,
	validTo string,
	prizes []prize.UpdatePrizeInput,
	updatedBy uuid.UUID,
	isActive bool,
) (*UpdatePrizeStructureOutput, error) {
	// Call the actual service
	input := prize.UpdatePrizeStructureInput{
		ID:          id,
		Name:        name,
		Description: description,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Prizes:      prizes,
		UpdatedBy:   updatedBy,
		IsActive:    isActive,
	}

	output, err := p.updatePrizeStructureService.UpdatePrizeStructure(input)
	if err != nil {
		return nil, err
	}

	// Convert prizes for response
	prizesOutput := make([]PrizeItem, 0, len(output.Prizes))
	for _, prize := range output.Prizes {
		prizesOutput = append(prizesOutput, PrizeItem{
			ID:                prize.ID,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}

	// Return response
	return &UpdatePrizeStructureOutput{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		StartDate:   output.StartDate,
		EndDate:     output.EndDate,
		Prizes:      prizesOutput,
		IsActive:    output.IsActive,
	}, nil
}

// DeletePrizeStructure deletes a prize structure
func (p *PrizeServiceAdapter) DeletePrizeStructure(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Call the actual service
	input := prize.DeletePrizeStructureInput{
		ID:        id,
		DeletedBy: deletedBy,
	}

	return p.deletePrizeStructureService.DeletePrizeStructure(input)
}
