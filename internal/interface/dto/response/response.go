package response

import (
	// No imports needed
)

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
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
	TotalEntries         int              `json:"totalEntries"`
	ExecutedByAdminID    string           `json:"executedByAdminID"`
	Winners              []WinnerResponse `json:"winners,omitempty"`
	CreatedAt            string           `json:"createdAt"`
	UpdatedAt            string           `json:"updatedAt,omitempty"`
}

// WinnerResponse represents a winner response
type WinnerResponse struct {
	ID            string `json:"id"`
	DrawID        string `json:"drawID,omitempty"`
	MSISDN        string `json:"msisdn"`
	PrizeTierID   string `json:"prizeTierID"`
	PrizeTierName string `json:"prizeTierName,omitempty"`
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

// EligibilityStatsResponse represents eligibility statistics response
type EligibilityStatsResponse struct {
	Date                 string `json:"date,omitempty"`
	TotalEligibleMSISDNs int    `json:"totalEligibleMSISDNs"`
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
}

// PrizeTierResponse represents a prize tier response
type PrizeTierResponse struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	PrizeType         string `json:"prizeType"`
	Value             string `json:"value"`
	ValueNGN          int    `json:"valueNGN,omitempty"`
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
	ID            string `json:"id"`
	UploadedBy    string `json:"uploadedBy"`
	UploadedAt    string `json:"uploadedAt"`
	TotalUploaded int    `json:"totalUploaded"`
	Status        string `json:"status"`
	Details       string `json:"details,omitempty"`
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
