package draw

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// InvokeRunnerUpService provides functionality for invoking runner-ups
type InvokeRunnerUpService struct {
	drawRepository draw.DrawRepository
	auditService   audit.AuditService
}

// NewInvokeRunnerUpService creates a new InvokeRunnerUpService
func NewInvokeRunnerUpService(
	drawRepository draw.DrawRepository,
	auditService audit.AuditService,
) *InvokeRunnerUpService {
	return &InvokeRunnerUpService{
		drawRepository: drawRepository,
		auditService:   auditService,
	}
}

// InvokeRunnerUpInput defines the input for the InvokeRunnerUp use case
type InvokeRunnerUpInput struct {
	WinnerID      uuid.UUID
	AdminUserID   uuid.UUID
	Reason        string
}

// InvokeRunnerUpOutput defines the output for the InvokeRunnerUp use case
type InvokeRunnerUpOutput struct {
	OriginalWinner RunnerUpWinnerOutput
	NewWinner      RunnerUpWinnerOutput
}

// RunnerUpWinnerOutput defines the winner output structure for runner-up invocation
type RunnerUpWinnerOutput struct {
	ID          uuid.UUID
	MSISDN      string
	PrizeTierID uuid.UUID
	PrizeValue  float64
	Status      string
}

// InvokeRunnerUp invokes a runner-up to replace a winner
func (uc *InvokeRunnerUpService) InvokeRunnerUp(ctx context.Context, input InvokeRunnerUpInput) (*InvokeRunnerUpOutput, error) {
	// Validate input
	if input.WinnerID == uuid.Nil {
		return nil, errors.New("winner ID is required")
	}
	
	if input.AdminUserID == uuid.Nil {
		return nil, errors.New("admin user ID is required")
	}
	
	// Get the original winner
	originalWinner, err := uc.drawRepository.GetWinnerByID(input.WinnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get winner: %w", err)
	}
	
	// Check if winner is eligible for replacement
	if originalWinner.IsRunnerUp {
		return nil, errors.New("cannot replace a runner-up")
	}
	
	// Get available runner-ups
	runnerUps, err := uc.drawRepository.GetRunnerUps(originalWinner.DrawID, originalWinner.PrizeTierID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get runner-ups: %w", err)
	}
	
	if len(runnerUps) == 0 {
		return nil, draw.NewDrawError(draw.ErrNoRunnerUpsAvailable, "No runner-ups available", nil)
	}
	
	// Select the first runner-up
	newWinner := runnerUps[0]
	
	// Update original winner status
	originalWinner.Status = "Replaced"
	originalWinner.UpdatedAt = time.Now()
	
	if err := uc.drawRepository.UpdateWinner(originalWinner); err != nil {
		return nil, fmt.Errorf("failed to update original winner: %w", err)
	}
	
	// Update runner-up status
	newWinner.IsRunnerUp = false
	newWinner.Status = "PendingNotification"
	newWinner.UpdatedAt = time.Now()
	
	if err := uc.drawRepository.UpdateWinner(&newWinner); err != nil {
		return nil, fmt.Errorf("failed to update runner-up: %w", err)
	}
	
	// Log audit
	if err := uc.auditService.LogAudit(
		"INVOKE_RUNNER_UP",
		"Winner",
		originalWinner.ID,
		input.AdminUserID,
		fmt.Sprintf("Runner-up invoked to replace winner %s", originalWinner.MSISDN),
		fmt.Sprintf("Reason: %s, New winner: %s", input.Reason, newWinner.MSISDN),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &InvokeRunnerUpOutput{
		OriginalWinner: RunnerUpWinnerOutput{
			ID:          originalWinner.ID,
			MSISDN:      originalWinner.MSISDN,
			PrizeTierID: originalWinner.PrizeTierID,
			PrizeValue:  originalWinner.PrizeValue,
			Status:      originalWinner.Status,
		},
		NewWinner: RunnerUpWinnerOutput{
			ID:          newWinner.ID,
			MSISDN:      newWinner.MSISDN,
			PrizeTierID: newWinner.PrizeTierID,
			PrizeValue:  newWinner.PrizeValue,
			Status:      newWinner.Status,
		},
	}, nil
}
