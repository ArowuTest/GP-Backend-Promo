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

// UpdatePrizeStructureService provides functionality for updating prize structures
type UpdatePrizeStructureService struct {
	prizeRepository prize.PrizeRepository
	auditService    audit.AuditService
}

// NewUpdatePrizeStructureService creates a new UpdatePrizeStructureService
func NewUpdatePrizeStructureService(
	prizeRepository prize.PrizeRepository,
	auditService audit.AuditService,
) *UpdatePrizeStructureService {
	return &UpdatePrizeStructureService{
		prizeRepository: prizeRepository,
		auditService:    auditService,
	}
}

// UpdatePrizeStructureInput defines the input for the UpdatePrizeStructure use case
type UpdatePrizeStructureInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   string // Format: YYYY-MM-DD
	EndDate     string // Format: YYYY-MM-DD
	Prizes      []UpdatePrizeInput
	UpdatedBy   uuid.UUID
}

// UpdatePrizeInput defines the input for a prize tier in update operation
type UpdatePrizeInput struct {
	ID          uuid.UUID // Optional, if not provided, a new prize will be created
	Name        string
	Description string
	Value       string
	Quantity    int
}

// UpdatePrizeStructureOutput defines the output for the UpdatePrizeStructure use case
type UpdatePrizeStructureOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Prizes      []UpdatePrizeOutput
}

// UpdatePrizeOutput defines the output for a prize tier in update operation
type UpdatePrizeOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Value       string
	Quantity    int
}

// UpdatePrizeStructure updates an existing prize structure
func (s *UpdatePrizeStructureService) UpdatePrizeStructure(ctx context.Context, input UpdatePrizeStructureInput) (*UpdatePrizeStructureOutput, error) {
	// Validate input
	if input.ID == uuid.Nil {
		return nil, errors.New("prize structure ID is required")
	}
	
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	
	if len(input.Prizes) == 0 {
		return nil, errors.New("at least one prize is required")
	}
	
	// Get existing prize structure
	existingPrizeStructure, err := s.prizeRepository.GetPrizeStructureByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prize structure: %w", err)
	}
	
	// Parse dates
	startDate, err := parseUpdateDate(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	
	endDate, err := parseUpdateDate(input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	
	// Update prize structure
	existingPrizeStructure.Name = input.Name
	existingPrizeStructure.Description = input.Description
	existingPrizeStructure.StartDate = startDate
	existingPrizeStructure.EndDate = endDate
	
	// Update prizes
	existingPrizeStructure.Prizes = make([]prize.PrizeTier, 0, len(input.Prizes))
	for i, prizeInput := range input.Prizes {
		prizeTier := prize.PrizeTier{
			PrizeStructureID: input.ID,
			Rank:             i + 1,
			Name:             prizeInput.Name,
			Description:      prizeInput.Description,
			Value:            prizeInput.Value,
			ValueNGN:         0, // Default value, can be calculated if needed
			Quantity:         prizeInput.Quantity,
		}
		
		if prizeInput.ID == uuid.Nil {
			// New prize
			prizeTier.ID = uuid.New()
		} else {
			// Existing prize
			prizeTier.ID = prizeInput.ID
		}
		
		existingPrizeStructure.Prizes = append(existingPrizeStructure.Prizes, prizeTier)
	}
	
	// Save prize structure
	if err := s.prizeRepository.UpdatePrizeStructure(existingPrizeStructure); err != nil {
		return nil, fmt.Errorf("failed to update prize structure: %w", err)
	}
	
	// Log audit
	if err := s.auditService.LogAudit(
		"UPDATE_PRIZE_STRUCTURE",
		"PrizeStructure",
		input.ID,
		input.UpdatedBy,
		fmt.Sprintf("Prize structure updated: %s", input.Name),
		fmt.Sprintf("Prizes: %d", len(input.Prizes)),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	// Prepare output
	prizeOutputs := make([]UpdatePrizeOutput, 0, len(existingPrizeStructure.Prizes))
	for _, prizeTier := range existingPrizeStructure.Prizes {
		prizeOutputs = append(prizeOutputs, UpdatePrizeOutput{
			ID:          prizeTier.ID,
			Name:        prizeTier.Name,
			Description: prizeTier.Description,
			Value:       prizeTier.Value,
			Quantity:    prizeTier.Quantity,
		})
	}
	
	return &UpdatePrizeStructureOutput{
		ID:          existingPrizeStructure.ID,
		Name:        existingPrizeStructure.Name,
		Description: existingPrizeStructure.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Prizes:      prizeOutputs,
	}, nil
}

// Helper function to parse date string
func parseUpdateDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
