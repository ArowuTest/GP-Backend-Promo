package prize

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// CreatePrizeStructureService provides functionality for creating prize structures
type CreatePrizeStructureService struct {
	prizeRepository prize.PrizeRepository
	auditService    audit.AuditService
}

// NewCreatePrizeStructureService creates a new CreatePrizeStructureService
func NewCreatePrizeStructureService(
	prizeRepository prize.PrizeRepository,
	auditService audit.AuditService,
) *CreatePrizeStructureService {
	return &CreatePrizeStructureService{
		prizeRepository: prizeRepository,
		auditService:    auditService,
	}
}

// CreatePrizeStructureInput defines the input for the CreatePrizeStructure use case
type CreatePrizeStructureInput struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	Prizes      []PrizeInput
	CreatedBy   uuid.UUID
	IsActive    bool
}

// PrizeInput defines the input for a prize tier
type PrizeInput struct {
	Name              string
	Description       string
	Value             float64
	Quantity          int
	NumberOfRunnerUps int
}

// CreatePrizeStructureOutput defines the output for the CreatePrizeStructure use case
type CreatePrizeStructureOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	Prizes      []CreatePrizeOutput
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsActive    bool
}

// CreatePrizeOutput defines the output for a prize tier in create operation
type CreatePrizeOutput struct {
	ID                uuid.UUID
	Name              string
	Description       string
	Value             float64
	Quantity          int
	NumberOfRunnerUps int
}

// CreatePrizeStructure creates a new prize structure
func (s *CreatePrizeStructureService) CreatePrizeStructure(ctx context.Context, input CreatePrizeStructureInput) (*CreatePrizeStructureOutput, error) {
	// Validate input
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	
	if len(input.Prizes) == 0 {
		return nil, errors.New("at least one prize is required")
	}
	
	// Create prize structure
	prizeStructureID := uuid.New()
	now := time.Now()
	
	prizeStructure := &prize.PrizeStructure{
		ID:          prizeStructureID,
		Name:        input.Name,
		Description: input.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    input.IsActive,
		Prizes:      make([]prize.PrizeTier, 0, len(input.Prizes)),
	}
	
	// Create prizes
	for i, prizeInput := range input.Prizes {
		prizeItem := prize.PrizeTier{
			ID:                prizeStructureID,
			PrizeStructureID:  prizeStructureID,
			Rank:              i + 1,
			Name:              prizeInput.Name,
			Description:       prizeInput.Description,
			Value:             prizeInput.Value,
			ValueNGN:          0, // Default value, can be calculated if needed
			Quantity:          prizeInput.Quantity,
			NumberOfRunnerUps: prizeInput.NumberOfRunnerUps,
		}	
		prizeStructure.Prizes = append(prizeStructure.Prizes, prizeItem)
	}
	
	// Save prize structure
	if err := s.prizeRepository.CreatePrizeStructure(prizeStructure); err != nil {
		return nil, fmt.Errorf("failed to create prize structure: %w", err)
	}
	
	// Log audit
	if err := s.auditService.LogAudit(
		"CREATE_PRIZE_STRUCTURE",
		"PrizeStructure",
		prizeStructureID,
		input.CreatedBy,
		fmt.Sprintf("Prize structure created: %s", input.Name),
		fmt.Sprintf("Prizes: %d", len(input.Prizes)),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	// Prepare output
	prizeOutputs := make([]CreatePrizeOutput, 0, len(prizeStructure.Prizes))
	for _, prizeTier := range prizeStructure.Prizes {
		prizeOutputs = append(prizeOutputs, CreatePrizeOutput{
			ID:                prizeTier.ID,
			Name:              prizeTier.Name,
			Description:       prizeTier.Description,
			Value:             prizeTier.Value,
			Quantity:          prizeTier.Quantity,
			NumberOfRunnerUps: prizeTier.NumberOfRunnerUps,
		})
	}
	
	return &CreatePrizeStructureOutput{
		ID:          prizeStructureID,
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Prizes:      prizeOutputs,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    input.IsActive,
	}, nil
}
