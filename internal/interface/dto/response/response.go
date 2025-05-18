package response

import (
	"time"
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
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalRows  int `json:"total_rows"`
	TotalPages int `json:"total_pages"`
}

// DrawResponse represents a draw response
type DrawResponse struct {
	ID                   string           `json:"id"`
	DrawDate             string           `json:"draw_date"`
	PrizeStructureID     string           `json:"prize_structure_id"`
	Status               string           `json:"status"`
	TotalEligibleMSISDNs int              `json:"total_eligible_msisdns"`
	TotalEntries         int              `json:"total_entries"`
	ExecutedByAdminID    string           `json:"executed_by_admin_id"`
	Winners              []WinnerResponse `json:"winners,omitempty"`
	CreatedAt            string           `json:"created_at"`
}

// WinnerResponse represents a winner response
type WinnerResponse struct {
	ID            string `json:"id"`
	MSISDN        string `json:"msisdn"`
	PrizeTierID   string `json:"prize_tier_id"`
	PrizeTierName string `json:"prize_tier_name,omitempty"`
	PrizeValue    string `json:"prize_value,omitempty"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status,omitempty"`
	PaymentNotes  string `json:"payment_notes,omitempty"`
	PaidAt        string `json:"paid_at,omitempty"`
	IsRunnerUp    bool   `json:"is_runner_up"`
	RunnerUpRank  int    `json:"runner_up_rank"`
}

// RunnerUpResponse represents a runner-up invocation response
type RunnerUpResponse struct {
	OriginalWinner WinnerResponse `json:"original_winner"`
	NewWinner      WinnerResponse `json:"new_winner"`
}

// EligibilityStatsResponse represents eligibility statistics response
type EligibilityStatsResponse struct {
	Date                 string `json:"date"`
	TotalEligibleMSISDNs int    `json:"total_eligible_msisdns"`
	TotalEntries         int    `json:"total_entries"`
}

// PrizeStructureResponse represents a prize structure response
type PrizeStructureResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	IsActive       bool                  `json:"is_active"`
	ValidFrom      string                `json:"valid_from"`
	ValidTo        string                `json:"valid_to,omitempty"`
	ApplicableDays []string              `json:"applicable_days,omitempty"`
	DayType        string                `json:"day_type,omitempty"`
	Prizes         []PrizeTierResponse   `json:"prizes"`
	CreatedAt      string                `json:"created_at,omitempty"`
	UpdatedAt      string                `json:"updated_at,omitempty"`
}

// PrizeTierResponse represents a prize tier response
type PrizeTierResponse struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	PrizeType         string `json:"prize_type"`
	Value             string `json:"value"`
	ValueNGN          int    `json:"valueNGN,omitempty"`
	Quantity          int    `json:"quantity"`
	Order             int    `json:"order"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
}

// ParticipantResponse represents a participant response
type ParticipantResponse struct {
	ID             string `json:"id"`
	MSISDN         string `json:"msisdn"`
	Points         int    `json:"points"`
	RechargeAmount float64 `json:"recharge_amount"`
	RechargeDate   string `json:"recharge_date"`
	CreatedAt      string `json:"created_at"`
}

// ParticipantStatsResponse represents participant statistics response
type ParticipantStatsResponse struct {
	Date              string `json:"date"`
	TotalParticipants int    `json:"total_participants"`
	TotalPoints       int    `json:"total_points"`
}

// UploadResponse represents a participant data upload response
type UploadResponse struct {
	AuditID           string   `json:"audit_id"`
	Status            string   `json:"status"`
	TotalRowsProcessed int      `json:"total_data_rows_processed"`
	SuccessfulRows    int      `json:"successful_rows_imported"`
	ErrorCount        int      `json:"errors_count"`
	ErrorDetails      []string `json:"errors"`
	DuplicatesSkipped int      `json:"duplicates_skipped_count,omitempty"`
}

// UploadAuditResponse represents an upload audit response
type UploadAuditResponse struct {
	ID              string   `json:"id"`
	UploadedBy      string   `json:"uploaded_by"`
	UploadDate      string   `json:"upload_date"`
	FileName        string   `json:"file_name"`
	Status          string   `json:"status"`
	TotalRows       int      `json:"total_rows"`
	SuccessfulRows  int      `json:"successful_rows"`
	ErrorCount      int      `json:"error_count"`
	ErrorDetails    []string `json:"error_details,omitempty"`
}

// AuditLogResponse represents an audit log response
type AuditLogResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	Timestamp string `json:"timestamp"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
	User      UserResponse `json:"user"`
}
