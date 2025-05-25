package response

// This file contains prize-related response types
// Some types are complementary to avoid duplication with response.go

import (
	"time"
)

// PrizeTierResponse represents a prize tier response
// This is used by the handler and must match the expected structure
type PrizeTierResponse struct {
	ID                string    `json:"id,omitempty"`
	PrizeStructureID  string    `json:"prize_structure_id,omitempty"`
	Rank              int       `json:"rank"`
	Name              string    `json:"name"`
	PrizeType         string    `json:"prizeType"`
	Description       string    `json:"description,omitempty"`
	Value             float64   `json:"value"`
	CurrencyCode      string    `json:"currency_code"`
	ValueNGN          float64   `json:"value_ngn,omitempty"`
	Quantity          int       `json:"quantity"`
	Order             int       `json:"order"`
	NumberOfRunnerUps int       `json:"numberOfRunnerUps"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// REMOVED: PrizeResponse is an alias for PrizeTierResponse to maintain compatibility
// This resolves the type mismatch in the handler
// type PrizeResponse = PrizeTierResponse

// PrizeTierDetailResponse represents a detailed view of a prize tier
// This is complementary to PrizeTierResponse
type PrizeTierDetailResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Value             float64   `json:"value"`
	CurrencyCode      string    `json:"currency_code"`
	Quantity          int       `json:"quantity"`
	NumberOfRunnerUps int       `json:"number_of_runner_ups"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// PrizeStructureDetailResponse represents a detailed view of a prize structure
// This is complementary to PrizeStructureResponse in response.go
type PrizeStructureDetailResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	StartDate   time.Time                `json:"start_date"`
	EndDate     time.Time                `json:"end_date"`
	IsActive    bool                     `json:"is_active"`
	DayType     string                   `json:"day_type"` // "weekday", "weekend", "all"
	PrizeTiers  []PrizeTierDetailResponse `json:"prize_tiers"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	CreatedBy   string                   `json:"created_by"`
	UpdatedBy   string                   `json:"updated_by"`
}

// PrizeStructureSummaryResponse represents a summarized view of a prize structure
// This is complementary to PrizeStructureResponse in response.go
type PrizeStructureSummaryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsActive    bool      `json:"is_active"`
	DayType     string    `json:"day_type"`
	TotalPrizes int       `json:"total_prizes"`
}

// CreatePrizeStructureRequest represents a request to create a prize structure
// This is used by the handler and must be preserved
type CreatePrizeStructureRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	IsActive    bool                    `json:"is_active"`
	ValidFrom   time.Time               `json:"valid_from" binding:"required"`
	ValidTo     *time.Time              `json:"valid_to"`
	Prizes      []CreatePrizeTierRequest `json:"prizes" binding:"required,dive"`
}

// CreatePrizeTierRequest represents a request to create a prize tier
// This is used by the handler and must be preserved
type CreatePrizeTierRequest struct {
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// UpdatePrizeStructureRequest represents a request to update a prize structure
// This is used by the handler and must be preserved
type UpdatePrizeStructureRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	IsActive    bool                    `json:"is_active"`
	ValidFrom   time.Time               `json:"valid_from" binding:"required"`
	ValidTo     *time.Time              `json:"valid_to"`
	Prizes      []UpdatePrizeTierRequest `json:"prizes" binding:"required,dive"`
}

// UpdatePrizeTierRequest represents a request to update a prize tier
// This is used by the handler and must be preserved
type UpdatePrizeTierRequest struct {
	ID                string  `json:"id"`
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// PrizeStructureCreateRequest represents the request to create a new prize structure
// This is a complementary type with strategic field naming
type PrizeStructureCreateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	StartDate   time.Time               `json:"start_date" binding:"required"`
	EndDate     time.Time               `json:"end_date" binding:"required"`
	IsActive    bool                    `json:"is_active"`
	DayType     string                  `json:"day_type" binding:"required,oneof=weekday weekend all"`
	PrizeTiers  []PrizeTierCreateRequest `json:"prize_tiers" binding:"required,dive"`
}

// PrizeTierCreateRequest represents the request to create a new prize tier
// This is a complementary type with strategic field naming
type PrizeTierCreateRequest struct {
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required,gt=0"`
	CurrencyCode      string  `json:"currency_code" binding:"required"`
	Quantity          int     `json:"quantity" binding:"required,gt=0"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups" binding:"gte=0"`
}

// PrizeStructureUpdateRequest represents the request to update an existing prize structure
// This is a complementary type with strategic field naming
type PrizeStructureUpdateRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	StartDate   time.Time               `json:"start_date"`
	EndDate     time.Time               `json:"end_date"`
	IsActive    bool                    `json:"is_active"`
	DayType     string                  `json:"day_type" binding:"omitempty,oneof=weekday weekend all"`
	PrizeTiers  []PrizeTierUpdateRequest `json:"prize_tiers" binding:"omitempty,dive"`
}

// PrizeTierUpdateRequest represents the request to update an existing prize tier
// This is a complementary type with strategic field naming
type PrizeTierUpdateRequest struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"omitempty,gt=0"`
	CurrencyCode      string  `json:"currency_code"`
	Quantity          int     `json:"quantity" binding:"omitempty,gt=0"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups" binding:"omitempty,gte=0"`
}
