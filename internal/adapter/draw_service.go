package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
)

// DrawServiceAdapter adapts the draw service to a consistent interface
type DrawServiceAdapter struct {
	executeDrawService       draw.ExecuteDrawService
	getDrawService           draw.GetDrawService
	listDrawsService         draw.ListDrawsService
	getEligibilityStatsService draw.GetEligibilityStatsService
	invokeRunnerUpService    draw.InvokeRunnerUpService
}

// NewDrawServiceAdapter creates a new DrawServiceAdapter
func NewDrawServiceAdapter(
	executeDrawService draw.ExecuteDrawService,
	getDrawService draw.GetDrawService,
	listDrawsService draw.ListDrawsService,
	getEligibilityStatsService draw.GetEligibilityStatsService,
	invokeRunnerUpService draw.InvokeRunnerUpService,
) *DrawServiceAdapter {
	return &DrawServiceAdapter{
		executeDrawService:       executeDrawService,
		getDrawService:           getDrawService,
		listDrawsService:         listDrawsService,
		getEligibilityStatsService: getEligibilityStatsService,
		invokeRunnerUpService:    invokeRunnerUpService,
	}
}

// Winner represents a draw winner
type Winner struct {
	ID          string
	MSISDN      string
	PrizeID     string
	PrizeName   string
	PrizeValue  float64
	IsInvoked   bool
	InvokedAt   *time.Time
	InvokedByID *string
}

// Draw represents a draw
type Draw struct {
	ID                  string
	Name                string
	Description         string
	DrawDate            time.Time
	PrizeStructureID    string
	PrizeStructureName  string
	TotalEligibleMSISDNs int
	TotalEntries        int
	ExecutedByAdminID   string
	ExecutedByAdminName string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Winners             []Winner
}

// ExecuteDrawOutput represents the output of ExecuteDraw
type ExecuteDrawOutput struct {
	Draw    Draw
	Winners []Winner
}

// GetDrawOutput represents the output of GetDraw
type GetDrawOutput struct {
	Draw Draw
}

// ListDrawsOutput represents the output of ListDraws
type ListDrawsOutput struct {
	Draws      []Draw
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

// GetEligibilityStatsOutput represents the output of GetEligibilityStats
type GetEligibilityStatsOutput struct {
	TotalEligibleMSISDNs int
	TotalEntries        int
	LastUpdated         time.Time
}

// InvokeRunnerUpOutput represents the output of InvokeRunnerUp
type InvokeRunnerUpOutput struct {
	Success   bool
	Winner    Winner
	RunnerUp  Winner
}

// ExecuteDraw executes a draw
func (d *DrawServiceAdapter) ExecuteDraw(
	ctx context.Context,
	name string,
	description string,
	drawDate time.Time,
	prizeStructureID uuid.UUID,
	executedByAdminID uuid.UUID,
) (*ExecuteDrawOutput, error) {
	// Call the actual service
	input := draw.ExecuteDrawInput{
		Name:             name,
		Description:      description,
		DrawDate:         drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedBy:       executedByAdminID,
	}

	output, err := d.executeDrawService.ExecuteDraw(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert winners for response
	winners := make([]Winner, 0, len(output.Winners))
	for _, w := range output.Winners {
		var invokedAt *time.Time
		var invokedByID *string
		if w.InvokedAt != nil {
			t := *w.InvokedAt
			invokedAt = &t
		}
		if w.InvokedByID != nil {
			id := w.InvokedByID.String()
			invokedByID = &id
		}

		winners = append(winners, Winner{
			ID:          w.ID.String(),
			MSISDN:      w.MSISDN,
			PrizeID:     w.PrizeID.String(),
			PrizeName:   w.PrizeName,
			PrizeValue:  w.PrizeValue,
			IsInvoked:   w.IsInvoked,
			InvokedAt:   invokedAt,
			InvokedByID: invokedByID,
		})
	}

	// Return response
	return &ExecuteDrawOutput{
		Draw: Draw{
			ID:                  output.ID.String(),
			Name:                output.Name,
			Description:         output.Description,
			DrawDate:            output.DrawDate,
			PrizeStructureID:    output.PrizeStructureID.String(),
			PrizeStructureName:  output.PrizeStructureName,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:        output.TotalEntries,
			ExecutedByAdminID:   output.ExecutedByAdminID.String(),
			ExecutedByAdminName: output.ExecutedByAdminName,
			CreatedAt:           output.CreatedAt,
			UpdatedAt:           output.UpdatedAt,
			Winners:             winners,
		},
		Winners: winners,
	}, nil
}

// GetDraw gets a draw by ID
func (d *DrawServiceAdapter) GetDraw(ctx context.Context, id uuid.UUID) (*GetDrawOutput, error) {
	// Call the actual service
	output, err := d.getDrawService.GetDraw(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert winners for response
	winners := make([]Winner, 0, len(output.Winners))
	for _, w := range output.Winners {
		var invokedAt *time.Time
		var invokedByID *string
		if w.InvokedAt != nil {
			t := *w.InvokedAt
			invokedAt = &t
		}
		if w.InvokedByID != nil {
			id := w.InvokedByID.String()
			invokedByID = &id
		}

		winners = append(winners, Winner{
			ID:          w.ID.String(),
			MSISDN:      w.MSISDN,
			PrizeID:     w.PrizeID.String(),
			PrizeName:   w.PrizeName,
			PrizeValue:  w.PrizeValue,
			IsInvoked:   w.IsInvoked,
			InvokedAt:   invokedAt,
			InvokedByID: invokedByID,
		})
	}

	// Return response
	return &GetDrawOutput{
		Draw: Draw{
			ID:                  output.ID.String(),
			Name:                output.Name,
			Description:         output.Description,
			DrawDate:            output.DrawDate,
			PrizeStructureID:    output.PrizeStructureID.String(),
			PrizeStructureName:  output.PrizeStructureName,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:        output.TotalEntries,
			ExecutedByAdminID:   output.ExecutedByAdminID.String(),
			ExecutedByAdminName: output.ExecutedByAdminName,
			CreatedAt:           output.CreatedAt,
			UpdatedAt:           output.UpdatedAt,
			Winners:             winners,
		},
	}, nil
}

// ListDraws lists draws with pagination
func (d *DrawServiceAdapter) ListDraws(ctx context.Context, page, pageSize int) (*ListDrawsOutput, error) {
	// Call the actual service
	input := draw.ListDrawsInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := d.listDrawsService.ListDraws(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert draws for response
	draws := make([]Draw, 0, len(output.Draws))
	for _, draw := range output.Draws {
		// Convert winners for response
		winners := make([]Winner, 0, len(draw.Winners))
		for _, w := range draw.Winners {
			var invokedAt *time.Time
			var invokedByID *string
			if w.InvokedAt != nil {
				t := *w.InvokedAt
				invokedAt = &t
			}
			if w.InvokedByID != nil {
				id := w.InvokedByID.String()
				invokedByID = &id
			}

			winners = append(winners, Winner{
				ID:          w.ID.String(),
				MSISDN:      w.MSISDN,
				PrizeID:     w.PrizeID.String(),
				PrizeName:   w.PrizeName,
				PrizeValue:  w.PrizeValue,
				IsInvoked:   w.IsInvoked,
				InvokedAt:   invokedAt,
				InvokedByID: invokedByID,
			})
		}

		draws = append(draws, Draw{
			ID:                  draw.ID.String(),
			Name:                draw.Name,
			Description:         draw.Description,
			DrawDate:            draw.DrawDate,
			PrizeStructureID:    draw.PrizeStructureID.String(),
			PrizeStructureName:  draw.PrizeStructureName,
			TotalEligibleMSISDNs: draw.TotalEligibleMSISDNs,
			TotalEntries:        draw.TotalEntries,
			ExecutedByAdminID:   draw.ExecutedByAdminID.String(),
			ExecutedByAdminName: draw.ExecutedByAdminName,
			CreatedAt:           draw.CreatedAt,
			UpdatedAt:           draw.UpdatedAt,
			Winners:             winners,
		})
	}

	// Return response
	return &ListDrawsOutput{
		Draws:      draws,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}, nil
}

// GetEligibilityStats gets eligibility stats
func (d *DrawServiceAdapter) GetEligibilityStats(ctx context.Context) (*GetEligibilityStatsOutput, error) {
	// Call the actual service
	output, err := d.getEligibilityStatsService.GetEligibilityStats(ctx)
	if err != nil {
		return nil, err
	}

	// Return response
	return &GetEligibilityStatsOutput{
		TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
		TotalEntries:        output.TotalEntries,
		LastUpdated:         output.LastUpdated,
	}, nil
}

// InvokeRunnerUp invokes a runner up
func (d *DrawServiceAdapter) InvokeRunnerUp(
	ctx context.Context,
	drawID uuid.UUID,
	winnerID uuid.UUID,
	runnerUpID uuid.UUID,
	invokedByAdminID uuid.UUID,
) (*InvokeRunnerUpOutput, error) {
	// Call the actual service
	input := draw.InvokeRunnerUpInput{
		DrawID:           drawID,
		WinnerID:         winnerID,
		RunnerUpID:       runnerUpID,
		InvokedByAdminID: invokedByAdminID,
	}

	output, err := d.invokeRunnerUpService.InvokeRunnerUp(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert winner for response
	var winnerInvokedAt *time.Time
	var winnerInvokedByID *string
	if output.Winner.InvokedAt != nil {
		t := *output.Winner.InvokedAt
		winnerInvokedAt = &t
	}
	if output.Winner.InvokedByID != nil {
		id := output.Winner.InvokedByID.String()
		winnerInvokedByID = &id
	}

	winner := Winner{
		ID:          output.Winner.ID.String(),
		MSISDN:      output.Winner.MSISDN,
		PrizeID:     output.Winner.PrizeID.String(),
		PrizeName:   output.Winner.PrizeName,
		PrizeValue:  output.Winner.PrizeValue,
		IsInvoked:   output.Winner.IsInvoked,
		InvokedAt:   winnerInvokedAt,
		InvokedByID: winnerInvokedByID,
	}

	// Convert runner up for response
	var runnerUpInvokedAt *time.Time
	var runnerUpInvokedByID *string
	if output.RunnerUp.InvokedAt != nil {
		t := *output.RunnerUp.InvokedAt
		runnerUpInvokedAt = &t
	}
	if output.RunnerUp.InvokedByID != nil {
		id := output.RunnerUp.InvokedByID.String()
		runnerUpInvokedByID = &id
	}

	runnerUp := Winner{
		ID:          output.RunnerUp.ID.String(),
		MSISDN:      output.RunnerUp.MSISDN,
		PrizeID:     output.RunnerUp.PrizeID.String(),
		PrizeName:   output.RunnerUp.PrizeName,
		PrizeValue:  output.RunnerUp.PrizeValue,
		IsInvoked:   output.RunnerUp.IsInvoked,
		InvokedAt:   runnerUpInvokedAt,
		InvokedByID: runnerUpInvokedByID,
	}

	// Return response
	return &InvokeRunnerUpOutput{
		Success:   output.Success,
		Winner:    winner,
		RunnerUp:  runnerUp,
	}, nil
}
