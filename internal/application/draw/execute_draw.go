package application

import (
	"time"

	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// ExecuteDrawUseCase represents the use case for executing a draw
type ExecuteDrawUseCase struct {
	drawRepository  draw.DrawRepository
	prizeRepository prize.PrizeRepository
}

// NewExecuteDrawUseCase creates a new ExecuteDrawUseCase
func NewExecuteDrawUseCase(
	drawRepository draw.DrawRepository,
	prizeRepository prize.PrizeRepository,
) *ExecuteDrawUseCase {
	return &ExecuteDrawUseCase{
		drawRepository:  drawRepository,
		prizeRepository: prizeRepository,
	}
}

// ExecuteDrawInput represents the input for the execute draw use case
type ExecuteDrawInput struct {
	DrawDate         time.Time
	PrizeStructureID uuid.UUID
	ExecutedByAdminID uuid.UUID
}

// ExecuteDrawOutput represents the output of the execute draw use case
type ExecuteDrawOutput struct {
	Draw     *draw.Draw
	Winners  []draw.Winner
	ErrorMsg string
}

// Execute executes the draw for the given date and prize structure
func (uc *ExecuteDrawUseCase) Execute(input ExecuteDrawInput) (*ExecuteDrawOutput, error) {
	// Validate input
	if err := draw.ValidateDrawDate(input.DrawDate); err != nil {
		return nil, draw.NewDrawError(draw.ErrInvalidDrawDate, "Invalid draw date", err)
	}

	// Check if draw already exists for this date
	existingDraw, err := uc.drawRepository.GetByDate(input.DrawDate)
	if err == nil && existingDraw != nil {
		return nil, draw.NewDrawError(draw.ErrDrawAlreadyExists, "Draw already exists for this date", nil)
	}

	// Get prize structure
	prizeStructure, err := uc.prizeRepository.GetPrizeStructureByID(input.PrizeStructureID)
	if err != nil {
		return nil, prize.NewPrizeError(prize.ErrPrizeStructureNotFound, "Prize structure not found", err)
	}

	// Get eligibility stats
	totalEligibleMSISDNs, totalEntries, err := uc.drawRepository.GetEligibilityStats(input.DrawDate)
	if err != nil {
		return nil, draw.NewDrawError(draw.ErrNoEligibleParticipants, "Failed to get eligibility stats", err)
	}

	if totalEligibleMSISDNs == 0 || totalEntries == 0 {
		return nil, draw.NewDrawError(draw.ErrNoEligibleParticipants, "No eligible participants for draw", nil)
	}

	// Create new draw
	newDraw := &draw.Draw{
		ID:                  uuid.New(),
		DrawDate:            input.DrawDate,
		PrizeStructureID:    input.PrizeStructureID,
		Status:              "Pending",
		TotalEligibleMSISDNs: totalEligibleMSISDNs,
		TotalEntries:        totalEntries,
		ExecutedByAdminID:   input.ExecutedByAdminID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save the draw
	if err := uc.drawRepository.Create(newDraw); err != nil {
		return nil, draw.NewDrawError("DRAW_CREATION_FAILED", "Failed to create draw", err)
	}

	// Execute the draw algorithm and select winners
	// This would involve complex logic to select winners based on points and eligibility
	// For each prize tier, select a winner and runner-ups

	// For demonstration purposes, we'll just create a placeholder for the output
	output := &ExecuteDrawOutput{
		Draw:     newDraw,
		Winners:  []draw.Winner{},
		ErrorMsg: "",
	}

	// Update draw status to completed
	newDraw.Status = "Completed"
	if err := uc.drawRepository.Update(newDraw); err != nil {
		return output, draw.NewDrawError("DRAW_UPDATE_FAILED", "Failed to update draw status", err)
	}

	return output, nil
}
