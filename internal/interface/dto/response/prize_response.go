package response

import (
	"time"

	"github.com/google/uuid"
)

// PrizeStructureResponse represents a prize structure response
type PrizeStructureResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsActive    bool                `json:"is_active"`
	ValidFrom   time.Time           `json:"valid_from"`
	ValidTo     *time.Time          `json:"valid_to"`
	Prizes      []PrizeTierResponse `json:"prizes"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// PrizeTierResponse represents a prize tier response
type PrizeTierResponse struct {
	ID                string    `json:"id"`
	PrizeStructureID  string    `json:"prize_structure_id"`
	Rank              int       `json:"rank"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Value             float64   `json:"value"`
	CurrencyCode      string    `json:"currency_code"` // Added currency code field
	ValueNGN          float64   `json:"value_ngn"`
	Quantity          int       `json:"quantity"`
	NumberOfRunnerUps int       `json:"number_of_runner_ups"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreatePrizeStructureRequest represents a request to create a prize structure
type CreatePrizeStructureRequest struct {
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description"`
	IsActive    bool                       `json:"is_active"`
	ValidFrom   time.Time                  `json:"valid_from" binding:"required"`
	ValidTo     *time.Time                 `json:"valid_to"`
	Prizes      []CreatePrizeTierRequest   `json:"prizes" binding:"required,dive"`
}

// CreatePrizeTierRequest represents a request to create a prize tier
type CreatePrizeTierRequest struct {
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Added currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// UpdatePrizeStructureRequest represents a request to update a prize structure
type UpdatePrizeStructureRequest struct {
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description"`
	IsActive    bool                       `json:"is_active"`
	ValidFrom   time.Time                  `json:"valid_from" binding:"required"`
	ValidTo     *time.Time                 `json:"valid_to"`
	Prizes      []UpdatePrizeTierRequest   `json:"prizes" binding:"required,dive"`
}

// UpdatePrizeTierRequest represents a request to update a prize tier
type UpdatePrizeTierRequest struct {
	ID                string  `json:"id"`
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Added currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}
