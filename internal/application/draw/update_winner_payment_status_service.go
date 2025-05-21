package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
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
	// Implementation using domain types
	winner, err := s.repository.UpdateWinnerPaymentStatus(ctx, input.WinnerID, input.PaymentStatus, input.Notes, input.UpdatedBy)
	if err != nil {
		return UpdateWinnerPaymentStatusOutput{}, err
	}
	
	// Convert paidAt to string format if needed
	paidAtStr := ""
	if winner.PaidAt != nil {
		paidAtStr = winner.PaidAt.Format("2006-01-02 15:04:05")
	}
	
	return UpdateWinnerPaymentStatusOutput{
		ID:            winner.ID,
		DrawID:        winner.DrawID,
		MSISDN:        winner.MSISDN,
		PrizeTierID:   winner.PrizeTierID,
		PrizeTierName: winner.PrizeTierName,
		PrizeValue:    winner.PrizeValue,
		Status:        winner.Status,
		PaymentStatus: winner.PaymentStatus,
		PaymentNotes:  winner.PaymentNotes,
		PaidAt:        paidAtStr,
		IsRunnerUp:    winner.IsRunnerUp,
		RunnerUpRank:  winner.RunnerUpRank,
		CreatedAt:     winner.CreatedAt,
		UpdatedAt:     winner.UpdatedAt,
	}, nil
}
