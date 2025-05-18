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
	StartDate   string // Format: YYYY-MM-DD
	EndDate     string // Format: YYYY-MM-DD
	Prizes      []PrizeInput
	CreatedBy   uuid.UUID
}

// PrizeInput defines the input for a prize tier
type PrizeInput struct {
	Name        string
	Description string
	Value       string
	Quantity    int
}

// CreatePrizeStructureOutput defines the output for the CreatePrizeStructure use case
type CreatePrizeStructureOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Prizes      []CreatePrizeOutput
}

// CreatePrizeOutput defines the output for a prize tier in create operation
type CreatePrizeOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Value       string
	Quantity    int
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
	
	// Parse dates
	startDate, err := parseDate(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	
	endDate, err := parseDate(input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	
	// Create prize structure
	prizeStructureID := uuid.New()
	prizeStructure := &prize.PrizeStructure{
		ID:          prizeStructureID,
		Name:        input.Name,
		Description: input.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedBy:   input.CreatedBy,
		Prizes:      make([]prize.PrizeTier, 0, len(input.Prizes)),
	}
	
	// Create prizes
	for i, prizeInput := range input.Prizes {
		prizeItem := prize.PrizeTier{
			ID:               uuid.New(),
			PrizeStructureID: prizeStructureID,
			Rank:             i + 1,
			Name:             prizeInput.Name,
			Description:      prizeInput.Description,
			Value:            prizeInput.Value,
			ValueNGN:         0, // Default value, can be calculated if needed
			Quantity:         prizeInput.Quantity,
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
			ID:          prizeTier.ID,
			Name:        prizeTier.Name,
			Description: prizeTier.Description,
			Value:       prizeTier.Value,
			Quantity:    prizeTier.Quantity,
		})
	}
	
	return &CreatePrizeStructureOutput{
		ID:          prizeStructureID,
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Prizes:      prizeOutputs,
	}, nil
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
