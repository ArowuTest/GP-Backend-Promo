package response

import (
	"github.com/google/uuid"
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

// PrizeStructureResponse defines the response for a prize structure
type PrizeStructureResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ValidFrom   string          `json:"validFrom"` // Format: YYYY-MM-DD
	ValidTo     string          `json:"validTo"`   // Format: YYYY-MM-DD
	Prizes      []PrizeResponse `json:"prizes"`
	IsActive    bool            `json:"isActive"`
}

// PrizeResponse defines the response for a prize tier
type PrizeResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Value             string    `json:"value"`
	Quantity          int       `json:"quantity"`
	NumberOfRunnerUps int       `json:"numberOfRunnerUps"`
}

// UserResponse defines the response for a user
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	IsActive bool      `json:"isActive"`
}

// LoginResponse defines the response for user login
type LoginResponse struct {
	Token  string       `json:"token"`
	User   UserResponse `json:"user"`
	Expiry string       `json:"expiry"`
}

// DrawResponse defines the response for a draw
type DrawResponse struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	DrawDate        string           `json:"drawDate"`
	Status          string           `json:"status"`
	PrizeStructure  string           `json:"prizeStructure"`
	Winners         []WinnerResponse `json:"winners"`
	CreatedAt       string           `json:"createdAt"`
	CreatedBy       string           `json:"createdBy"`
}

// WinnerResponse defines the response for a winner
type WinnerResponse struct {
	ID            uuid.UUID `json:"id"`
	MSISDN        string    `json:"msisdn"`
	MaskedMSISDN  string    `json:"maskedMsisdn"`
	PrizeName     string    `json:"prizeName"`
	PrizeValue    string    `json:"prizeValue"`
	PaymentStatus string    `json:"paymentStatus"`
	PaymentDate   string    `json:"paymentDate"`
	PaymentRef    string    `json:"paymentRef"`
	IsRunnerUp    bool      `json:"isRunnerUp"`
	InvokedAt     string    `json:"invokedAt"`
}

// AuditLogResponse defines the response for an audit log
type AuditLogResponse struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"userId"`
	Username    string                 `json:"username"`
	Action      string                 `json:"action"`
	EntityType  string                 `json:"entityType"`
	EntityID    string                 `json:"entityId"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	IPAddress   string                 `json:"ipAddress"`
	UserAgent   string                 `json:"userAgent"`
	CreatedAt   string                 `json:"createdAt"`
}

// UploadAuditResponse defines the response for an upload audit
type UploadAuditResponse struct {
	ID          uuid.UUID `json:"id"`
	FileName    string    `json:"fileName"`
	RecordCount int       `json:"recordCount"`
	UploadedBy  string    `json:"uploadedBy"`
	UploadedAt  string    `json:"uploadedAt"`
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
	TotalEligible int `json:"totalEligible"`
	TotalEntries  int `json:"totalEntries"`
}
