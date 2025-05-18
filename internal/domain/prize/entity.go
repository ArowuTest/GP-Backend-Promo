package prize

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PrizeStructure represents a prize structure entity in the domain
type PrizeStructure struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsActive    bool
	ValidFrom   time.Time
	ValidTo     *time.Time
	StartDate   time.Time
	EndDate     time.Time
	CreatedBy   uuid.UUID
	Prizes      []PrizeTier
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Prize represents a prize tier within a prize structure
type Prize struct {
	ID               uuid.UUID
	PrizeStructureID uuid.UUID
	Name             string
	Description      string
	Value            string // Monetary value as string to support different formats
	Quantity         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PrizeTier represents a prize tier within a prize structure
type PrizeTier struct {
	ID               uuid.UUID
	PrizeStructureID uuid.UUID
	Rank             int
	Name             string
	Description      string
	Value            string // Monetary value as string to support different formats
	ValueNGN         float64
	Quantity         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PrizeRepository defines the interface for prize structure data access
type PrizeRepository interface {
	CreatePrizeStructure(prizeStructure *PrizeStructure) error
	GetPrizeStructureByID(id uuid.UUID) (*PrizeStructure, error)
	ListPrizeStructures(page, pageSize int) ([]PrizeStructure, int, error)
	UpdatePrizeStructure(prizeStructure *PrizeStructure) error
	DeletePrizeStructure(id uuid.UUID) error
	GetActivePrizeStructure(date time.Time) (*PrizeStructure, error)
	
	CreatePrize(prize *Prize) error
	GetPrizeByID(id uuid.UUID) (*Prize, error)
	ListPrizesByStructureID(structureID uuid.UUID) ([]Prize, error)
	UpdatePrize(prize *Prize) error
	DeletePrize(id uuid.UUID) error
	
	CreatePrizeTier(prizeTier *PrizeTier) error
	GetPrizeTierByID(id uuid.UUID) (*PrizeTier, error)
	ListPrizeTiersByStructureID(structureID uuid.UUID) ([]PrizeTier, error)
	UpdatePrizeTier(prizeTier *PrizeTier) error
	DeletePrizeTier(id uuid.UUID) error
}

// PrizeError represents domain-specific errors for the prize domain
type PrizeError struct {
	Code    string
	Message string
	Err     error
}

// Error codes for the prize domain
const (
	ErrPrizeStructureNotFound = "PRIZE_STRUCTURE_NOT_FOUND"
	ErrPrizeNotFound          = "PRIZE_NOT_FOUND"
	ErrPrizeTierNotFound      = "PRIZE_TIER_NOT_FOUND"
	ErrInvalidPrizeStructure  = "INVALID_PRIZE_STRUCTURE"
	ErrInvalidPrize           = "INVALID_PRIZE"
	ErrInvalidPrizeTier       = "INVALID_PRIZE_TIER"
	ErrNoPrizeStructureActive = "NO_PRIZE_STRUCTURE_ACTIVE"
	ErrInvalidDateRange       = "INVALID_DATE_RANGE"
)

// Error implements the error interface
func (e *PrizeError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *PrizeError) Unwrap() error {
	return e.Err
}

// NewPrizeError creates a new PrizeError
func NewPrizeError(code, message string, err error) *PrizeError {
	return &PrizeError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidatePrizeStructure validates that a prize structure is valid
func ValidatePrizeStructure(ps *PrizeStructure) error {
	if ps.Name == "" {
		return errors.New("prize structure name cannot be empty")
	}
	
	if ps.ValidFrom.IsZero() {
		return errors.New("valid from date cannot be empty")
	}
	
	if ps.ValidTo != nil && ps.ValidTo.Before(ps.ValidFrom) {
		return errors.New("valid to date must be after valid from date")
	}
	
	if len(ps.Prizes) == 0 {
		return errors.New("prize structure must have at least one prize")
	}
	
	return nil
}

// ValidatePrize validates that a prize is valid
func ValidatePrize(p *Prize) error {
	if p.Name == "" {
		return errors.New("prize name cannot be empty")
	}
	
	if p.Quantity < 1 {
		return errors.New("prize quantity must be positive")
	}
	
	return nil
}

// ValidatePrizeTier validates that a prize tier is valid
func ValidatePrizeTier(pt *PrizeTier) error {
	if pt.Name == "" {
		return errors.New("prize tier name cannot be empty")
	}
	
	if pt.Quantity < 1 {
		return errors.New("prize tier quantity must be positive")
	}
	
	return nil
}
