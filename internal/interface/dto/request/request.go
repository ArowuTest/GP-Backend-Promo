package request

import (
	"time"
)

// ExecuteDrawRequest represents the request to execute a draw
type ExecuteDrawRequest struct {
	DrawDate         string `json:"drawDate" binding:"required"`
	PrizeStructureID string `json:"prizeStructureId" binding:"required"`
}

// InvokeRunnerUpRequest represents the request to invoke a runner-up
type InvokeRunnerUpRequest struct {
	WinnerID string `json:"winnerId" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

// UpdateWinnerPaymentStatusRequest represents the request to update a winner's payment status
type UpdateWinnerPaymentStatusRequest struct {
	PaymentStatus string `json:"paymentStatus" binding:"required"`
	Notes         string `json:"notes"`
}

// CreatePrizeStructureRequest represents the request to create a prize structure
type CreatePrizeStructureRequest struct {
	Name           string                `json:"name" binding:"required"`
	Description    string                `json:"description" binding:"required"`
	IsActive       bool                  `json:"isActive"`
	ValidFrom      string                `json:"validFrom" binding:"required"`
	ValidTo        *string               `json:"validTo"`
	ApplicableDays []string              `json:"applicableDays"`
	Prizes         []CreatePrizeTierRequest `json:"prizes" binding:"required,dive"`
}

// CreatePrizeTierRequest represents the request to create a prize tier
type CreatePrizeTierRequest struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name" binding:"required"`
	PrizeType         string `json:"prizeType" binding:"required"`
	Value             string `json:"value" binding:"required"`
	Quantity          int    `json:"quantity" binding:"required"`
	Order             int    `json:"order" binding:"required"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
}

// UpdatePrizeStructureRequest represents the request to update a prize structure
type UpdatePrizeStructureRequest struct {
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	IsActive       *bool                 `json:"isActive"`
	ValidFrom      string                `json:"validFrom"`
	ValidTo        *string               `json:"validTo"`
	ApplicableDays []string              `json:"applicableDays"`
	Prizes         []CreatePrizeTierRequest `json:"prizes"`
}

// LoginRequest represents the request to login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"` // Added Email field to support login by email
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

// GetAuditLogsRequest represents the request to get audit logs
type GetAuditLogsRequest struct {
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	UserID    string `form:"userId"`
	Action    string `form:"action"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"pageSize,default=10"`
}

// UploadParticipantsRequest represents the request to upload participants
type UploadParticipantsRequest struct {
	Participants []ParticipantInput `json:"participants" binding:"required,dive"`
}

// ParticipantInput represents a participant input for upload
type ParticipantInput struct {
	MSISDN         string  `json:"msisdn" binding:"required"`
	RechargeAmount float64 `json:"rechargeAmount" binding:"required"`
	RechargeDate   string  `json:"rechargeDate" binding:"required"`
}

// ParseTime parses a date string in YYYY-MM-DD format to time.Time
func ParseTime(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
