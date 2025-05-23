package request

import (
	"time"

	"github.com/google/uuid"
)

// CreatePrizeStructureRequest defines the request for creating a prize structure
type CreatePrizeStructureRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	ValidFrom   string                `json:"validFrom" binding:"required"` // Format: YYYY-MM-DD
	ValidTo     string                `json:"validTo"`                      // Format: YYYY-MM-DD
	Prizes      []CreatePrizeRequest  `json:"prizes" binding:"required,dive"`
	IsActive    bool                  `json:"isActive"`
}

// CreatePrizeRequest defines the request for creating a prize tier
type CreatePrizeRequest struct {
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	Value             string `json:"value" binding:"required"`
	Quantity          int    `json:"quantity" binding:"required,min=1"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps" binding:"min=0"`
}

// UpdatePrizeStructureRequest defines the request for updating a prize structure
type UpdatePrizeStructureRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	ValidFrom   string                `json:"validFrom" binding:"required"` // Format: YYYY-MM-DD
	ValidTo     string                `json:"validTo"`                      // Format: YYYY-MM-DD
	Prizes      []UpdatePrizeRequest  `json:"prizes" binding:"required,dive"`
	IsActive    bool                  `json:"isActive"`
}

// UpdatePrizeRequest defines the request for updating a prize tier
type UpdatePrizeRequest struct {
	ID                string `json:"id"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	Value             string `json:"value" binding:"required"`
	Quantity          int    `json:"quantity" binding:"required,min=1"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps" binding:"min=0"`
}

// LoginRequest defines the request for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest defines the request for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRequest defines the request for updating a user
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"omitempty,min=8"`
	Role     string `json:"role"`
	IsActive bool   `json:"isActive"`
}

// ResetPasswordRequest defines the request for resetting a user's password
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// UploadParticipantsRequest defines the request for uploading participants
type UploadParticipantsRequest struct {
	FileName string `json:"fileName" binding:"required"`
	Data     string `json:"data" binding:"required"` // Base64 encoded CSV data
}

// ExecuteDrawRequest defines the request for executing a draw
type ExecuteDrawRequest struct {
	Name            string    `json:"name" binding:"required"`
	Description     string    `json:"description"`
	DrawDate        time.Time `json:"drawDate" binding:"required"`
	PrizeStructureID uuid.UUID `json:"prizeStructureId" binding:"required"`
}

// UpdateWinnerPaymentStatusRequest defines the request for updating a winner's payment status
type UpdateWinnerPaymentStatusRequest struct {
	PaymentStatus string `json:"paymentStatus" binding:"required"`
	PaymentDate   string `json:"paymentDate"`
	PaymentRef    string `json:"paymentRef"`
}

// InvokeRunnerUpRequest defines the request for invoking a runner-up
type InvokeRunnerUpRequest struct {
	Reason string `json:"reason" binding:"required"`
}
