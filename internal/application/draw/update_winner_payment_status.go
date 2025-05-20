package draw

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// UpdateWinnerPaymentStatusService provides functionality for updating a winner's payment status
type UpdateWinnerPaymentStatusService struct {
	drawRepository drawDomain.DrawRepository
}

// NewUpdateWinnerPaymentStatusService creates a new UpdateWinnerPaymentStatusService
func NewUpdateWinnerPaymentStatusService(drawRepository drawDomain.DrawRepository) *UpdateWinnerPaymentStatusService {
	return &UpdateWinnerPaymentStatusService{
		drawRepository: drawRepository,
	}
}

// UpdateWinnerPaymentStatusInput defines the input for the UpdateWinnerPaymentStatus use case
type UpdateWinnerPaymentStatusInput struct {
	WinnerID      uuid.UUID
	PaymentStatus string
	Notes         string
	UpdatedBy     uuid.UUID
}

// UpdateWinnerPaymentStatusOutput defines the output for the UpdateWinnerPaymentStatus use case
type UpdateWinnerPaymentStatusOutput struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	PrizeTierName string
	PrizeValue    string
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        time.Time
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UpdateWinnerPaymentStatus updates a winner's payment status
func (s *UpdateWinnerPaymentStatusService) UpdateWinnerPaymentStatus(ctx context.Context, input UpdateWinnerPaymentStatusInput) (UpdateWinnerPaymentStatusOutput, error) {
	// Get current winner
	winner, err := s.drawRepository.GetWinnerByID(ctx, input.WinnerID)
	if err != nil {
		return UpdateWinnerPaymentStatusOutput{}, fmt.Errorf("failed to get winner: %w", err)
	}
	
	// Update payment status
	winner.PaymentStatus = input.PaymentStatus
	winner.PaymentNotes = input.Notes
	winner.UpdatedAt = time.Now()
	
	// Set paid date if status is PAID
	if input.PaymentStatus == "PAID" {
		now := time.Now()
		winner.PaidAt = &now
	}
	
	// Save updated winner
	err = s.drawRepository.UpdateWinner(ctx, *winner)
	if err != nil {
		return UpdateWinnerPaymentStatusOutput{}, fmt.Errorf("failed to update winner: %w", err)
	}
	
	// Get updated winner
	updatedWinner, err := s.drawRepository.GetWinnerByID(ctx, input.WinnerID)
	if err != nil {
		return UpdateWinnerPaymentStatusOutput{}, fmt.Errorf("failed to get updated winner: %w", err)
	}
	
	// Create output
	paidAt := time.Time{}
	if updatedWinner.PaidAt != nil {
		paidAt = *updatedWinner.PaidAt
	}
	
	return UpdateWinnerPaymentStatusOutput{
		ID:            updatedWinner.ID,
		DrawID:        updatedWinner.DrawID,
		MSISDN:        updatedWinner.MSISDN,
		PrizeTierID:   updatedWinner.PrizeTierID,
		PrizeTierName: updatedWinner.PrizeTierName,
		PrizeValue:    updatedWinner.PrizeValue,
		Status:        updatedWinner.Status,
		PaymentStatus: updatedWinner.PaymentStatus,
		PaymentNotes:  updatedWinner.PaymentNotes,
		PaidAt:        paidAt,
		IsRunnerUp:    updatedWinner.IsRunnerUp,
		RunnerUpRank:  updatedWinner.RunnerUpRank,
		CreatedAt:     updatedWinner.CreatedAt,
		UpdatedAt:     updatedWinner.UpdatedAt,
	}, nil
}
