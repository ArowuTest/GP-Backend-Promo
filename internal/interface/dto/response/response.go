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
	Name                 string           `json:"name,omitempty"`
	Description          string           `json:"description,omitempty"`
	DrawDate             string           `json:"drawDate"`
	PrizeStructureID     string           `json:"prizeStructureID"`
	PrizeStructure       interface{}      `json:"prizeStructure,omitempty"`
	Status               string           `json:"status"`
	TotalEligibleMSISDNs int              `json:"totalEligibleMSISDNs"`
	TotalEntries         int              `json:"totalEntries"`
	ExecutedByAdminID    string           `json:"executedByAdminID"`
	CreatedBy            string           `json:"createdBy,omitempty"`
	Winners              []WinnerResponse `json:"winners,omitempty"`
	CreatedAt            string           `json:"createdAt"`
	UpdatedAt            string           `json:"updatedAt,omitempty"`
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
	Message         string         `json:"message"`
	OriginalWinner  WinnerResponse `json:"forfeited_winner"`
	NewWinner       WinnerResponse `json:"promoted_runner_up"`
}

// RunnerUpInvocationResult represents a runner-up invocation result
// This is an alias for RunnerUpResponse to maintain backward compatibility
type RunnerUpInvocationResult = RunnerUpResponse

// EligibilityStatsResponse represents eligibility statistics response
type EligibilityStatsResponse struct {
	Date                 string `json:"date,omitempty"`
	TotalEligibleMSISDNs int    `json:"totalEligibleMSISDNs"`
	TotalEntries         int    `json:"totalEntries"`
	TotalEligible        int    `json:"totalEligible,omitempty"` // Alias for TotalEligibleMSISDNs for backward compatibility
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
	UpdatedBy      string                `json:"updatedBy,omitempty"`
}

// PrizeTierResponse represents a prize tier response
type PrizeTierResponse struct {
	ID                string  `json:"id,omitempty"`
	PrizeStructureID  string  `json:"prize_structure_id,omitempty"`
	Name              string  `json:"name"`
	PrizeType         string  `json:"prizeType"`
	Description       string  `json:"description,omitempty"`
	Value             string  `json:"value"`
	ValueNGN          int     `json:"valueNGN,omitempty"`
	CurrencyCode      string  `json:"currency_code,omitempty"`
	Quantity          int     `json:"quantity"`
	Order             int     `json:"order"`
	Rank              int     `json:"rank,omitempty"`
	NumberOfRunnerUps int     `json:"numberOfRunnerUps"`
}

// ParticipantResponse represents a participant response
type ParticipantResponse struct {
	ID             string  `json:"id"`
	MSISDN         string  `json:"msisdn"`
	Points         int     `json:"points"`
	RechargeAmount float64 `json:"rechargeAmount"`
	RechargeDate   string  `json:"rechargeDate"`
	CreatedAt      string  `json:"createdAt"`
	UploadID       string  `json:"uploadID,omitempty"`
	UploadedAt     string  `json:"uploadedAt,omitempty"`
}

// ParticipantStatsResponse represents participant statistics response
type ParticipantStatsResponse struct {
	Date              string `json:"date"`
	TotalParticipants int    `json:"totalParticipants"`
	TotalPoints       int    `json:"totalPoints"`
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
	UploadedBy          string `json:"uploadedByUserId,omitempty"`
	UploadedAt          string `json:"uploadTimestamp,omitempty"`
	FileName            string `json:"fileName,omitempty"`
	TotalUploaded       int    `json:"recordCount,omitempty"`
	SuccessfullyImported int   `json:"successfullyImported,omitempty"`
	SuccessfulRows      int    `json:"successfulRows,omitempty"`
	DuplicatesSkipped   int    `json:"duplicatesSkipped,omitempty"`
	ErrorsEncountered   int    `json:"errorsEncountered,omitempty"`
	ErrorCount          int    `json:"errorCount,omitempty"`
	OperationType       string `json:"operationType,omitempty"`
	Status              string `json:"status"`
	Details             string `json:"notes,omitempty"`
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
	IsActive  bool   `json:"isActive"`
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

// PrizeTierDetailResponse represents a detailed view of a prize tier
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
// with additional fields for admin purposes
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
// for list views and dropdowns
type PrizeStructureSummaryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsActive    bool      `json:"is_active"`
	DayType     string    `json:"day_type"`
	TotalPrizes int       `json:"total_prizes"`
	PrizeCount  int       `json:"prizeCount,omitempty"`
	ValidFrom   string    `json:"validFrom,omitempty"`
	ValidTo     string    `json:"validTo,omitempty"`
	CreatedAt   string    `json:"createdAt,omitempty"`
}

// CreatePrizeStructureRequest represents the request to create a new prize structure
type CreatePrizeStructureRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	StartDate   time.Time               `json:"start_date" binding:"required"`
	EndDate     time.Time               `json:"end_date" binding:"required"`
	IsActive    bool                    `json:"is_active"`
	ValidFrom   time.Time               `json:"valid_from" binding:"required"`
	ValidTo     *time.Time              `json:"valid_to"`
	DayType     string                  `json:"day_type" binding:"required,oneof=weekday weekend all"`
	PrizeTiers  []PrizeTierCreateRequest `json:"prize_tiers" binding:"required,dive"`
	Prizes      []CreatePrizeTierRequest `json:"prizes" binding:"required,dive"`
}

// PrizeTierCreateRequest represents the request to create a new prize tier
type PrizeTierCreateRequest struct {
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required,gt=0"`
	CurrencyCode      string  `json:"currency_code" binding:"required"`
	Quantity          int     `json:"quantity" binding:"required,gt=0"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups" binding:"gte=0"`
}

// CreatePrizeTierRequest represents a request to create a prize tier
type CreatePrizeTierRequest struct {
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` 
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// PrizeStructureUpdateRequest represents the request to update an existing prize structure
type PrizeStructureUpdateRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	StartDate   time.Time               `json:"start_date"`
	EndDate     time.Time               `json:"end_date"`
	IsActive    bool                    `json:"is_active"`
	ValidFrom   time.Time               `json:"valid_from" binding:"required"`
	ValidTo     *time.Time              `json:"valid_to"`
	DayType     string                  `json:"day_type" binding:"omitempty,oneof=weekday weekend all"`
	PrizeTiers  []PrizeTierUpdateRequest `json:"prize_tiers" binding:"omitempty,dive"`
	Prizes      []UpdatePrizeTierRequest `json:"prizes" binding:"required,dive"`
}

// PrizeTierUpdateRequest represents the request to update an existing prize tier
type PrizeTierUpdateRequest struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"omitempty,gt=0"`
	CurrencyCode      string  `json:"currency_code"`
	Quantity          int     `json:"quantity" binding:"omitempty,gt=0"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups" binding:"omitempty,gte=0"`
}

// UpdatePrizeTierRequest represents a request to update a prize tier
type UpdatePrizeTierRequest struct {
	ID                string  `json:"id"`
	Rank              int     `json:"rank" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	Value             float64 `json:"value" binding:"required"`
	CurrencyCode      string  `json:"currency_code" binding:"required"` 
	Quantity          int     `json:"quantity" binding:"required"`
	NumberOfRunnerUps int     `json:"number_of_runner_ups"`
}

// DeleteConfirmationResponse represents a response for delete operations
type DeleteConfirmationResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
	Message string `json:"message,omitempty"`
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
