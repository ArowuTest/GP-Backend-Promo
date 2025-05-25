package response

import (
	"github.com/google/uuid"
	"time"
)

// SuccessResponse defines a standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse defines a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Pagination defines standard pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalRows  int   `json:"totalRows"`
	TotalPages int   `json:"totalPages"`
	TotalItems int64 `json:"totalItems"`
}

// PaginatedResponse defines a standard paginated response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// PrizeStructureResponse defines the response for a prize structure
type PrizeStructureResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ValidFrom   string          `json:"validFrom"` // Format: YYYY-MM-DD
	ValidTo     string          `json:"validTo"`   // Format: YYYY-MM-DD
	Prizes      []PrizeResponse `json:"prizes"`
	IsActive    bool            `json:"isActive"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

// PrizeResponse defines the response for a prize tier
// Updated to include all fields used by the handler
type PrizeResponse struct {
	ID               string    `json:"id"`
	PrizeStructureID string    `json:"prize_structure_id,omitempty"`
	Rank             int       `json:"rank"`
	Name             string    `json:"name"`
	PrizeType        string    `json:"prizeType"`
	Description      string    `json:"description,omitempty"`
	Value            float64   `json:"value"`
	CurrencyCode     string    `json:"currency_code"`
	ValueNGN         float64   `json:"value_ngn,omitempty"`
	Quantity         int       `json:"quantity"`
	Order            int       `json:"order"`
	NumberOfRunnerUps int      `json:"numberOfRunnerUps"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

// UserResponseBase defines the base response for a user (used internally)
type UserResponseBase struct {
	ID       uuid.UUID `json:"-"`
	Email    string    `json:"-"`
	Username string    `json:"-"`
	Role     string    `json:"-"`
	IsActive bool      `json:"-"`
}

// LoginResponse defines the response for user login
type LoginResponse struct {
	Token  string       `json:"token"`
	User   UserResponse `json:"user"`
	Expiry string       `json:"expiry"`
}

// DrawResponse defines the response for a draw
type DrawResponse struct {
	ID            uuid.UUID        `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	DrawDate      string           `json:"drawDate"`
	Status        string           `json:"status"`
	PrizeStructure string          `json:"prizeStructure"`
	Winners       []WinnerResponse `json:"winners"`
	CreatedAt     string           `json:"createdAt"`
	CreatedBy     string           `json:"createdBy"`
}

// WinnerResponse defines the response for a winner
type WinnerResponse struct {
	ID            uuid.UUID `json:"id"`
	DrawID        string    `json:"drawId"`           // Added to match frontend expectations
	MSISDN        string    `json:"msisdn"`
	MaskedMSISDN  string    `json:"maskedMsisdn"`
	PrizeTierID   string    `json:"prizeTierId"`      // Added to match frontend expectations
	PrizeName     string    `json:"prizeName"`
	PrizeValue    string    `json:"prizeValue"`
	PaymentStatus string    `json:"paymentStatus"`
	PaymentDate   string    `json:"paymentDate"`
	PaymentRef    string    `json:"paymentRef"`
	PaymentNotes  string    `json:"paymentNotes"`     // Added to match frontend expectations
	Status        string    `json:"status"`           // Added to match frontend expectations
	IsRunnerUp    bool      `json:"isRunnerUp"`
	RunnerUpRank  int       `json:"runnerUpRank"`     // Added to match frontend expectations
	InvokedAt     string    `json:"invokedAt"`
	CreatedAt     string    `json:"createdAt"`        // Added to match frontend expectations
	UpdatedAt     string    `json:"updatedAt"`        // Added to match frontend expectations
}

// AuditLogResponse defines the response for an audit log
type AuditLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Summary    string `json:"summary"`
	Details    string `json:"details"`
	CreatedAt  string `json:"createdAt"`
}

// DataUploadAuditResponse defines the response for a data upload audit
type DataUploadAuditResponse struct {
	ID                   string `json:"id"`
	UploadedBy           string `json:"uploadedBy"`
	UploadedAt           string `json:"uploadedAt"`
	FileName             string `json:"fileName"`
	TotalUploaded        int    `json:"totalUploaded"`
	SuccessfullyImported int    `json:"successfullyImported"`
	DuplicatesSkipped    int    `json:"duplicatesSkipped"`
	ErrorsEncountered    int    `json:"errorsEncountered"`
	Status               string `json:"status"`
	Details              string `json:"details"`
	OperationType        string `json:"operationType"`
}

// ParticipantResponse defines the response for a participant
type ParticipantResponse struct {
	ID        uuid.UUID `json:"id"`
	MSISDN    string    `json:"msisdn"`
	Points    int       `json:"points"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

// ParticipantStatsResponse defines the response for participant statistics
type ParticipantStatsResponse struct {
	TotalParticipants int     `json:"totalParticipants"`
	TotalPoints       int     `json:"totalPoints"`
	AveragePoints     float64 `json:"averagePoints"`
}

// EligibilityStatsResponse defines the response for eligibility statistics
type EligibilityStatsResponse struct {
	Date          string `json:"date,omitempty"` // Added to match frontend expectations
	TotalEligible int    `json:"totalEligible"`
	TotalEntries  int    `json:"totalEntries"`
}

// RunnerUpInvocationResult defines the response for invoking a runner-up
type RunnerUpInvocationResult struct {
	Message        string        `json:"message"`
	OriginalWinner WinnerResponse `json:"originalWinner"`
	NewWinner      WinnerResponse `json:"newWinner"`
}

// UserResponse defines the response for a user
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
