package draw

import (
	"context"
	"time"
	"github.com/google/uuid"
)

// UpdateWinnerPaymentStatusInput represents input for UpdateWinnerPaymentStatus
type UpdateWinnerPaymentStatusInput struct {
	WinnerID      string
	PaymentStatus string
	PaymentNotes  string
	Notes         string // Added for handler compatibility
	UpdatedBy     uuid.UUID // Added for handler compatibility
}

// UpdateWinnerPaymentStatusOutput represents output for UpdateWinnerPaymentStatus
type UpdateWinnerPaymentStatusOutput struct {
	Success       bool
	ID            uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	Status        string
	PaymentStatus string
	PaymentNotes  string
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
func (s *UpdateWinnerPaymentStatusService) UpdateWinnerPaymentStatus(ctx context.Context, input UpdateWinnerPaymentStatusInput) (*UpdateWinnerPaymentStatusOutput, error) {
	// Parse winner ID
	winnerID, err := uuid.Parse(input.WinnerID)
	if err != nil {
		return nil, err
	}

	// Get winner from repository
	winner, err := s.repository.GetWinnerByID(winnerID)
	if err != nil {
		return nil, err
	}

	// Update payment status
	winner.PaymentStatus = input.PaymentStatus
	
	// Use Notes if PaymentNotes is empty (for backward compatibility)
	if input.PaymentNotes != "" {
		winner.PaymentNotes = input.PaymentNotes
	} else {
		winner.PaymentNotes = input.Notes
	}

	// Save changes
	err = s.repository.UpdateWinner(winner)
	if err != nil {
		return nil, err
	}

	return &UpdateWinnerPaymentStatusOutput{
		Success:       true,
		ID:            winner.ID,
		MSISDN:        winner.MSISDN,
		PrizeTierID:   winner.PrizeTierID,
		Status:        winner.Status,
		PaymentStatus: winner.PaymentStatus,
		PaymentNotes:  winner.PaymentNotes,
		IsRunnerUp:    winner.IsRunnerUp,
		RunnerUpRank:  winner.RunnerUpRank,
		CreatedAt:     winner.CreatedAt,
		UpdatedAt:     time.Now(),
	}, nil
}
