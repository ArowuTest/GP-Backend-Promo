package draw

import (
	"context"
	"errors"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// InvokeRunnerUpInput represents the input for the InvokeRunnerUp use case
type InvokeRunnerUpInput struct {
	WinnerID  string
	Reason    string
	InvokedBy string
}

// InvokeRunnerUpOutput represents the output from the InvokeRunnerUp use case
type InvokeRunnerUpOutput struct {
	NewWinner draw.Winner
	OldWinner draw.Winner
}

// InvokeRunnerUpUseCase defines the use case for invoking a runner-up
type InvokeRunnerUpUseCase struct {
	drawRepo draw.Repository
}

// NewInvokeRunnerUpUseCase creates a new InvokeRunnerUpUseCase
func NewInvokeRunnerUpUseCase(drawRepo draw.Repository) *InvokeRunnerUpUseCase {
	return &InvokeRunnerUpUseCase{
		drawRepo: drawRepo,
	}
}

// Execute performs the invoke runner-up use case
func (uc *InvokeRunnerUpUseCase) Execute(ctx context.Context, input InvokeRunnerUpInput) (InvokeRunnerUpOutput, error) {
	// Validate input
	if input.WinnerID == "" {
		return InvokeRunnerUpOutput{}, draw.ErrInvalidWinnerID
	}
	if input.Reason == "" {
		return InvokeRunnerUpOutput{}, errors.New("reason is required")
	}
	if input.InvokedBy == "" {
		return InvokeRunnerUpOutput{}, errors.New("invoker information is required")
	}

	// Get the winner
	winner, err := uc.drawRepo.GetWinnerByID(ctx, input.WinnerID)
	if err != nil {
		return InvokeRunnerUpOutput{}, err
	}

	// Check if winner is already replaced
	if winner.Status == draw.WinnerStatusReplaced {
		return InvokeRunnerUpOutput{}, errors.New("winner has already been replaced")
	}

	// Get the next runner-up for this prize
	runnerUp, err := uc.drawRepo.GetNextRunnerUp(ctx, winner.DrawID, winner.PrizeID)
	if err != nil {
		return InvokeRunnerUpOutput{}, err
	}

	// Create a new winner from the runner-up
	newWinner := draw.Winner{
		DrawID:       winner.DrawID,
		PrizeID:      winner.PrizeID,
		MSISDN:       runnerUp.MSISDN,
		PrizeName:    winner.PrizeName,
		PrizeAmount:  winner.PrizeAmount,
		Status:       draw.WinnerStatusPending,
		SelectedAt:   time.Now(),
		PaymentStatus: draw.PaymentStatusPending,
	}

	// Update the old winner status
	winner.Status = draw.WinnerStatusReplaced
	winner.ReplacementReason = input.Reason
	winner.ReplacedBy = input.InvokedBy
	winner.ReplacedAt = time.Now()

	// Update the runner-up status
	runnerUp.Status = draw.RunnerUpStatusSelected

	// Perform the transaction
	err = uc.drawRepo.InvokeRunnerUp(ctx, winner, newWinner, runnerUp)
	if err != nil {
		return InvokeRunnerUpOutput{}, err
	}

	return InvokeRunnerUpOutput{
		NewWinner: newWinner,
		OldWinner: winner,
	}, nil
}
