package response

import (
	"time"
)

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalRows  int   `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
}

// DrawResponse represents a draw response
type DrawResponse struct {
	ID                   string           `json:"id"`
	DrawDate             string           `json:"drawDate"`
	PrizeStructureID     string           `json:"prizeStructureID"`
	Status               string           `json:"status"`
	TotalEligibleMSISDNs int              `json:"totalEligibleMSISDNs"`
	TotalEligible        int              `json:"totalEligible,omitempty"` // Alias for TotalEligibleMSISDNs
	TotalEntries         int              `json:"totalEntries"`
	ExecutedByAdminID    string           `json:"executedByAdminID"`
	Winners              []WinnerResponse `json:"winners,omitempty"`
	CreatedAt            string           `json:"createdAt"`
	UpdatedAt            string           `json:"updatedAt,omitempty"`
	Name                 string           `json:"name,omitempty"`
	Description          string           `json:"description,omitempty"`
	PrizeStructure       string           `json:"prizeStructure,omitempty"`
	CreatedBy            string           `json:"createdBy,omitempty"`
}

// WinnerResponse represents a winner response
type WinnerResponse struct {
	ID            string `json:"id"`
	DrawID        string `json:"drawID,omitempty"`
	MSISDN        string `json:"msisdn"`
	MaskedMSISDN  string `json:"maskedMSISDN,omitempty"`
	PrizeTierID   string `json:"prizeTierID"`
	PrizeTierName string `json:"prizeTierName,omitempty"`
	PrizeName     string `json:"prizeName,omitempty"`
	PrizeValue    string `json:"prizeValue,omitempty"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus,omitempty"`
	PaymentNotes  string `json:"paymentNotes,omitempty"`
	PaidAt        string `json:"paidAt,omitempty"`
	IsRunnerUp    bool   `json:"isRunnerUp"`
	RunnerUpRank  int    `json:"runnerUpRank"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// RunnerUpResponse represents a runner-up invocation response
type RunnerUpResponse struct {
	Message         string        `json:"message"`
	OriginalWinner  WinnerResponse `json:"forfeited_winner"`
	NewWinner       WinnerResponse `json:"promoted_runner_up"`
}

// RunnerUpInvocationResult is an alias for RunnerUpResponse
type RunnerUpInvocationResult RunnerUpResponse

// EligibilityStatsResponse represents eligibility statistics response
type EligibilityStatsResponse struct {
	Date                 string `json:"date,omitempty"`
	TotalEligibleMSISDNs int    `json:"totalEligibleMSISDNs"`
	TotalEligible        int    `json:"totalEligible,omitempty"` // Alias for TotalEligibleMSISDNs
	TotalEntries         int    `json:"totalEntries"`
}

// PrizeStructureResponse represents a prize structure response
type PrizeStructureResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	IsActive       bool                  `json:"isActive"`
	ValidFrom      string                `json:"validFrom"`
	ValidTo        string                `json:"validTo,omitempty"`
	ApplicableDays []string              `json:"applicableDays,omitempty"`
	DayType        string                `json:"dayType,omitempty"`
	Prizes         []PrizeTierResponse   `json:"prizes"`
	CreatedAt      string                `json:"createdAt,omitempty"`
	UpdatedAt      string                `json:"updatedAt,omitempty"`
	CreatedBy      string                `json:"createdBy,omitempty"`
}

// PrizeTierResponse represents a prize tier response
type PrizeTierResponse struct {
	ID                string `json:"id,omitempty"`
	PrizeStructureID  string `json:"prize_structure_id,omitempty"`
	Rank              int    `json:"rank,omitempty"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	PrizeType         string `json:"prizeType"`
	Value             string `json:"value"`
	ValueNGN          int    `json:"valueNGN,omitempty"`
	CurrencyCode      string `json:"currencyCode,omitempty"`
	Quantity          int    `json:"quantity"`
	Order             int    `json:"order"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
}

// ParticipantResponse represents a participant response
type ParticipantResponse struct {
	ID             string  `json:"id"`
	MSISDN         string  `json:"msisdn"`
	Points         int     `json:"points"`
	RechargeAmount float64 `json:"rechargeAmount"`
	RechargeDate   string  `json:"rechargeDate"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt,omitempty"`
	UploadID       string  `json:"uploadID,omitempty"`
	UploadedAt     string  `json:"uploadedAt,omitempty"`
}

// ParticipantStatsResponse represents participant statistics response
type ParticipantStatsResponse struct {
	Date              string  `json:"date"`
	TotalParticipants int     `json:"totalParticipants"`
	TotalPoints       int     `json:"totalPoints"`
	AveragePoints     float64 `json:"averagePoints,omitempty"`
}

// UploadResponse represents a participant data upload response
type UploadResponse struct {
	AuditID           string   `json:"auditID"`
	Status            string   `json:"status"`
	TotalRowsProcessed int      `json:"totalDataRowsProcessed"`
	SuccessfulRows    int      `json:"successfulRowsImported"`
	ErrorCount        int      `json:"errorsCount"`
	ErrorDetails      []string `json:"errors"`
	DuplicatesSkipped int      `json:"duplicatesSkippedCount,omitempty"`
}

// UploadAuditResponse represents an upload audit response
type UploadAuditResponse struct {
	ID              string   `json:"id"`
	UploadedBy      string   `json:"uploadedBy"`
	UploadDate      string   `json:"uploadDate"`
	FileName        string   `json:"fileName"`
	Status          string   `json:"status"`
	TotalRows       int      `json:"totalRows"`
	SuccessfulRows  int      `json:"successfulRows"`
	ErrorCount      int      `json:"errorCount"`
	ErrorDetails    []string `json:"errorDetails,omitempty"`
	UploadedAt      string   `json:"uploadedAt,omitempty"`
	TotalUploaded   int      `json:"totalUploaded,omitempty"`
}

// DataUploadAuditResponse represents a data upload audit response for reports
type DataUploadAuditResponse struct {
	ID                  string `json:"id"`
	UploadedBy          string `json:"uploadedBy"`
	UploadedByUserId    string `json:"uploadedByUserId,omitempty"`
	UploadedAt          string `json:"uploadedAt"`
	TotalUploaded       int    `json:"totalUploaded"`
	Status              string `json:"status"`
	Details             string `json:"details,omitempty"`
	FileName            string `json:"fileName,omitempty"`
	SuccessfullyImported int    `json:"successfullyImported,omitempty"`
	DuplicatesSkipped   int    `json:"duplicatesSkipped,omitempty"`
	ErrorsEncountered   int    `json:"errorsEncountered,omitempty"`
	OperationType       string `json:"operationType,omitempty"`
}

// AuditLogResponse represents an audit log response
type AuditLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"userID"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityID"`
	Summary    string `json:"summary"`
	Details    string `json:"details"`
	CreatedAt  string `json:"createdAt"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expiresAt"`
	User      UserResponse `json:"user"`
}

// UploadParticipantsResponse represents a participant upload response
type UploadParticipantsResponse struct {
	TotalUploaded int    `json:"totalUploaded"`
	UploadID      string `json:"uploadID"`
	UploadedAt    string `json:"uploadedAt"`
}

// DeleteConfirmationResponse represents a delete confirmation response
type DeleteConfirmationResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
	Entity  string `json:"entity,omitempty"`
}

// PrizeTierDetailResponse represents a detailed prize tier response
// with additional fields for admin views
type PrizeTierDetailResponse struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	PrizeType         string `json:"prizeType"`
	Value             string `json:"value"`
	ValueNGN          int    `json:"valueNGN,omitempty"`
	CurrencyCode      string `json:"currencyCode,omitempty"`
	Quantity          int    `json:"quantity"`
	Order             int    `json:"order"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
	Description       string `json:"description,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
}

// PrizeStructureDetailResponse represents a detailed prize structure response
// with additional fields for admin views
type PrizeStructureDetailResponse struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	IsActive       bool                    `json:"isActive"`
	ValidFrom      string                  `json:"validFrom"`
	ValidTo        string                  `json:"validTo,omitempty"`
	ApplicableDays []string                `json:"applicableDays,omitempty"`
	DayType        string                  `json:"dayType,omitempty"`
	Prizes         []PrizeTierDetailResponse `json:"prizes"`
	CreatedAt      string                  `json:"createdAt,omitempty"`
	UpdatedAt      string                  `json:"updatedAt,omitempty"`
	CreatedBy      string                  `json:"createdBy,omitempty"`
	UpdatedBy      string                  `json:"updatedBy,omitempty"`
}

// PrizeStructureSummaryResponse represents a summarized prize structure response
// for list views
type PrizeStructureSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
	ValidFrom   string `json:"validFrom"`
	ValidTo     string `json:"validTo,omitempty"`
	DayType     string `json:"dayType,omitempty"`
	PrizeCount  int    `json:"prizeCount"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

// CreatePrizeStructureRequest represents a request to create a prize structure
type CreatePrizeStructureRequest struct {
	Name           string                     `json:"name" binding:"required"`
	Description    string                     `json:"description"`
	IsActive       bool                       `json:"is_active"`
	ValidFrom      string                     `json:"valid_from" binding:"required"`
	ValidTo        *string                    `json:"valid_to"`
	ApplicableDays []string                   `json:"applicable_days"`
	Prizes         []CreatePrizeTierRequest   `json:"prizes" binding:"required,dive"`
}

// CreatePrizeTierRequest represents a request to create a prize tier
type CreatePrizeTierRequest struct {
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             string  `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Added currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// UpdatePrizeStructureRequest represents a request to update a prize structure
type UpdatePrizeStructureRequest struct {
	Name           string                     `json:"name" binding:"required"`
	Description    string                     `json:"description"`
	IsActive       bool                       `json:"is_active"`
	ValidFrom      string                     `json:"valid_from" binding:"required"`
	ValidTo        *string                    `json:"valid_to"`
	ApplicableDays []string                   `json:"applicable_days"`
	Prizes         []UpdatePrizeTierRequest   `json:"prizes" binding:"required,dive"`
}

// UpdatePrizeTierRequest represents a request to update a prize tier
type UpdatePrizeTierRequest struct {
	ID                string  `json:"id"`
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             string  `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` // Added currency code field
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}
