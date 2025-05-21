package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// UpdateWinnerPaymentStatusInput represents input for UpdateWinnerPaymentStatus
type UpdateWinnerPaymentStatusInput struct {
	WinnerID      uuid.UUID
	PaymentStatus string
	Notes         string
	UpdatedBy     uuid.UUID
}

// UpdateWinnerPaymentStatusOutput represents output for UpdateWinnerPaymentStatus
type UpdateWinnerPaymentStatusOutput struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	PrizeTierName string
	PrizeValue    float64
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        string
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UpdateWinnerPaymentStatusService handles updating winner payment status
type UpdateWinnerPaymentStatusService struct {
	repository Repository
}

// NewUpdateWinnerPaymentStatusService creates a new UpdateWinnerPaymentStatusService
func NewUpdateWinnerPaymentStatusService(repository Repository) *UpdateWinnerPaymentStatusService {
	return &UpdateWinnerPaymentStatusService{
		repository: repository,
	}
}

// UpdateWinnerPaymentStatus updates a winner's payment status
func (s *UpdateWinnerPaymentStatusService) UpdateWinnerPaymentStatus(ctx context.Context, input UpdateWinnerPaymentStatusInput) (UpdateWinnerPaymentStatusOutput, error) {
	// For now, return mock data
	paidAt := ""
	if input.PaymentStatus == "Paid" {
		paidAt = time.Now().Format("2006-01-02 15:04:05")
	}
	
	return UpdateWinnerPaymentStatusOutput{
		ID:            input.WinnerID,
		DrawID:        uuid.New(),
		MSISDN:        "234*****789",
		PrizeTierID:   uuid.New(),
		PrizeTierName: "First Prize",
		PrizeValue:    1000000,
		Status:        "Active",
		PaymentStatus: input.PaymentStatus,
		PaymentNotes:  input.Notes,
		PaidAt:        paidAt,
		IsRunnerUp:    false,
		RunnerUpRank:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}
