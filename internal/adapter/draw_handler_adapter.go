package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// DrawHandlerAdapter adapts the draw service adapter to match the handler's expected interface
type DrawHandlerAdapter struct {
	drawServiceAdapter *DrawServiceAdapter
}

// NewDrawHandlerAdapter creates a new DrawHandlerAdapter
func NewDrawHandlerAdapter(
	drawServiceAdapter *DrawServiceAdapter,
) *DrawHandlerAdapter {
	return &DrawHandlerAdapter{
		drawServiceAdapter: drawServiceAdapter,
	}
}

// ListDraws adapts GetDraws to match the handler's expected method name
func (a *DrawHandlerAdapter) ListDraws(
	ctx context.Context,
	page, pageSize int,
) (*entity.PaginatedDraws, error) {
	return a.drawServiceAdapter.ListDraws(ctx, page, pageSize)
}

// GetDrawByID adapts the service adapter's GetDrawByID to match the handler's expected output format
func (a *DrawHandlerAdapter) GetDrawByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.DrawWithWinners, error) {
	draw, err := a.drawServiceAdapter.GetDrawByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a wrapper that includes the Draw field expected by the handler
	return &entity.DrawWithWinners{
		Draw:    *draw,
		Winners: draw.Winners,
	}, nil
}

// ListWinners adapts GetWinners to match the handler's expected method signature
func (a *DrawHandlerAdapter) ListWinners(
	ctx context.Context,
	page, pageSize int,
	startDateStr, endDateStr string,
) (*entity.PaginatedWinners, error) {
	// Parse the draw ID if provided
	var drawID uuid.UUID
	
	// Default values for other filters
	msisdn := ""
	status := ""
	paymentStatus := ""
	isRunnerUp := false

	return a.drawServiceAdapter.GetWinners(ctx, page, pageSize, drawID, msisdn, status, paymentStatus, isRunnerUp)
}

// ExecuteDraw adapts the service adapter's ExecuteDraw to match the handler's expected signature
func (a *DrawHandlerAdapter) ExecuteDraw(
	ctx context.Context,
	drawDateStr string,
	prizeStructureID uuid.UUID,
	executedBy uuid.UUID,
) (*entity.Draw, error) {
	// Parse the draw date
	drawDate, err := time.Parse("2006-01-02", drawDateStr)
	if err != nil {
		return nil, err
	}

	// Default runner up count
	runnerUpCount := 3

	return a.drawServiceAdapter.ExecuteDraw(ctx, drawDate, prizeStructureID, executedBy, runnerUpCount)
}

// GetEligibilityStats adapts the service adapter's GetEligibilityStats to match the handler's expected signature
func (a *DrawHandlerAdapter) GetEligibilityStats(
	ctx context.Context,
	drawDateStr string,
) (*entity.EligibilityStats, error) {
	// Parse the draw date
	drawDate, err := time.Parse("2006-01-02", drawDateStr)
	if err != nil {
		return nil, err
	}

	return a.drawServiceAdapter.GetEligibilityStats(ctx, drawDate)
}

// InvokeRunnerUp adapts the service adapter's InvokeRunnerUp to match the handler's expected signature
func (a *DrawHandlerAdapter) InvokeRunnerUp(
	ctx context.Context,
	winnerID uuid.UUID,
	invokedBy uuid.UUID,
	reason string,
) (*entity.RunnerUpInvocationResult, error) {
	winner, err := a.drawServiceAdapter.InvokeRunnerUp(ctx, winnerID, reason, invokedBy)
	if err != nil {
		return nil, err
	}

	// Create a mock result since we don't have the original winner in the adapter response
	result := &entity.RunnerUpInvocationResult{
		NewWinner: *winner,
		OriginalWinner: entity.Winner{
			ID:          uuid.New(), // Placeholder
			Status:      "Forfeited",
			IsRunnerUp:  false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	return result, nil
}

// UpdateWinnerPaymentStatus adapts the service adapter's UpdateWinnerPaymentStatus to match the handler's expected signature
func (a *DrawHandlerAdapter) UpdateWinnerPaymentStatus(
	ctx context.Context,
	winnerIDStr string,
	paymentStatus string,
	paymentRef string,
	updatedBy uuid.UUID,
) (*entity.Winner, error) {
	// Parse the winner ID
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		return nil, err
	}

	return a.drawServiceAdapter.UpdateWinnerPaymentStatus(ctx, winnerID, paymentStatus, updatedBy)
}
