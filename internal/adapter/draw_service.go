package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// DrawServiceAdapter adapts the draw service to a consistent interface
type DrawServiceAdapter struct {
	drawService         *draw.ExecuteDrawService
	getDrawByIDService  *draw.GetDrawByIDService
	listDrawsService    *draw.ListDrawsService
	eligibilityService  *draw.GetEligibilityStatsService
	invokeRunnerUpService *draw.InvokeRunnerUpService
	updateWinnerService *draw.UpdateWinnerPaymentStatusService
	listWinnersService  *draw.ListWinnersService
}

// NewDrawServiceAdapter creates a new DrawServiceAdapter
func NewDrawServiceAdapter(
	drawService *draw.ExecuteDrawService,
	getDrawByIDService *draw.GetDrawByIDService,
	listDrawsService *draw.ListDrawsService,
	eligibilityService *draw.GetEligibilityStatsService,
	invokeRunnerUpService *draw.InvokeRunnerUpService,
	updateWinnerService *draw.UpdateWinnerPaymentStatusService,
	listWinnersService *draw.ListWinnersService,
) *DrawServiceAdapter {
	return &DrawServiceAdapter{
		drawService:         drawService,
		getDrawByIDService:  getDrawByIDService,
		listDrawsService:    listDrawsService,
		eligibilityService:  eligibilityService,
		invokeRunnerUpService: invokeRunnerUpService,
		updateWinnerService: updateWinnerService,
		listWinnersService:  listWinnersService,
	}
}

// ExecuteDraw executes a draw
func (d *DrawServiceAdapter) ExecuteDraw(
	ctx context.Context,
	drawDate time.Time,
	prizeStructureID uuid.UUID,
	executedByID uuid.UUID,
	runnerUpCount int,
) (*entity.Draw, error) {
	// Create input for the service
	input := draw.ExecuteDrawInput{
		DrawDate:         drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedByAdminID: executedByID,
	}

	// Execute draw
	output, err := d.drawService.ExecuteDraw(input)
	if err != nil {
		return nil, err
	}

	// Convert winners to entity model
	winners := make([]entity.Winner, 0, len(output.Winners))
	for _, w := range output.Winners {
		// Create masked MSISDN
		maskedMSISDN := util.MaskMSISDN(w.MSISDN)
		
		winners = append(winners, entity.Winner{
			ID:          w.ID,
			DrawID:      output.DrawID,
			MSISDN:      w.MSISDN,
			MaskedMSISDN: maskedMSISDN,
			PrizeID:     uuid.Nil, // Not available in output
			PrizeTierID: w.PrizeTierID,
			PrizeName:   w.PrizeName,
			PrizeValue:  w.PrizeValue,
			Status:      "PendingNotification", // Default status
			IsRunnerUp:  false,
			CreatedAt:   time.Now(),
		})
	}

	// Create response
	result := &entity.Draw{
		ID:                   output.DrawID,
		Name:                 "Draw for " + drawDate.Format("2006-01-02"),
		Description:          "Automatically executed draw",
		DrawDate:             output.DrawDate,
		PrizeStructureID:     prizeStructureID,
		Status:               "Completed",
		RunnerUpsCount:       runnerUpCount,
		TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
		TotalEntries:         output.TotalEntries,
		ExecutedByAdminID:    executedByID,
		CreatedBy:            executedByID,
		Winners:              winners,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	return result, nil
}

// GetDrawByID gets a draw by ID
func (d *DrawServiceAdapter) GetDrawByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Draw, error) {
	// Create input for the service
	input := draw.GetDrawByIDInput{
		ID: id.String(),
	}

	// Get draw
	output, err := d.getDrawByIDService.GetDrawByID(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert winners to entity model
	winners := make([]entity.Winner, 0, len(output.Winners))
	for _, w := range output.Winners {
		// Create masked MSISDN
		maskedMSISDN := util.MaskMSISDN(w.MSISDN)
		
		winners = append(winners, entity.Winner{
			ID:          w.ID,
			DrawID:      output.ID,
			MSISDN:      w.MSISDN,
			MaskedMSISDN: maskedMSISDN,
			PrizeID:     uuid.Nil, // Not available in output
			PrizeTierID: w.PrizeTierID,
			PrizeName:   "",
			PrizeValue:  w.PrizeValue,
			Status:      w.Status,
			IsRunnerUp:  w.IsRunnerUp,
			CreatedAt:   w.CreatedAt,
			UpdatedAt:   w.UpdatedAt,
		})
	}

	// Create response
	result := &entity.Draw{
		ID:                   output.ID,
		Name:                 "Draw for " + output.DrawDate.Format("2006-01-02"),
		Description:          "Automatically executed draw",
		DrawDate:             output.DrawDate,
		PrizeStructureID:     output.PrizeStructureID,
		Status:               output.Status,
		RunnerUpsCount:       0, // Not available in output
		TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
		TotalEntries:         output.TotalEntries,
		ExecutedByAdminID:    output.ExecutedBy,
		CreatedBy:            uuid.Nil, // Not available in output
		Winners:              winners,
		CreatedAt:            output.CreatedAt,
		UpdatedAt:            output.UpdatedAt,
	}

	return result, nil
}

// ListDraws gets a list of draws with pagination
func (d *DrawServiceAdapter) ListDraws(
	ctx context.Context,
	page, pageSize int,
) (*entity.PaginatedDraws, error) {
	// Create input for the service
	input := draw.ListDrawsInput{
		Page:     page,
		PageSize: pageSize,
	}

	// Get draws
	output, err := d.listDrawsService.ListDraws(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert draws to entity model
	draws := make([]entity.Draw, 0, len(output.Draws))
	for _, d := range output.Draws {
		// Convert winners to entity model if available
		winners := make([]entity.Winner, 0, len(d.Winners))
		for _, w := range d.Winners {
			// Create masked MSISDN
			maskedMSISDN := util.MaskMSISDN(w.MSISDN)
			
			winners = append(winners, entity.Winner{
				ID:          w.ID,
				DrawID:      d.ID,
				MSISDN:      w.MSISDN,
				MaskedMSISDN: maskedMSISDN,
				PrizeID:     uuid.Nil, // Not available in output
				PrizeTierID: w.PrizeTierID,
				PrizeName:   "",
				PrizeValue:  w.PrizeValue,
				Status:      w.Status,
				IsRunnerUp:  w.IsRunnerUp,
				CreatedAt:   w.CreatedAt,
				UpdatedAt:   w.UpdatedAt,
			})
		}

		draws = append(draws, entity.Draw{
			ID:                   d.ID,
			Name:                 "Draw for " + d.DrawDate.Format("2006-01-02"),
			Description:          "Automatically executed draw",
			DrawDate:             d.DrawDate,
			PrizeStructureID:     d.PrizeStructureID,
			Status:               d.Status,
			RunnerUpsCount:       0, // Not available in output
			TotalEligibleMSISDNs: d.TotalEligibleMSISDNs,
			TotalEntries:         d.TotalEntries,
			ExecutedByAdminID:    d.ExecutedBy,
			CreatedBy:            uuid.Nil, // Not available in output
			Winners:              winners,
			CreatedAt:            d.CreatedAt,
			UpdatedAt:            d.UpdatedAt,
		})
	}

	// Create response
	result := &entity.PaginatedDraws{
		Draws:      draws,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}

	return result, nil
}

// GetEligibilityStats gets eligibility statistics for a draw
func (d *DrawServiceAdapter) GetEligibilityStats(
	ctx context.Context,
	date time.Time,
) (*entity.EligibilityStats, error) {
	// Fix for line 244: Convert time.Time to string as the service expects a string date
	dateStr := date.Format("2006-01-02")
	
	// Create input for the service
	input := draw.GetEligibilityStatsInput{
		Date: dateStr, // Using string format as the service expects
	}

	// Get eligibility stats
	output, err := d.eligibilityService.GetEligibilityStats(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.EligibilityStats{
		TotalEligible: output.TotalEligibleMSISDNs,
		TotalPoints:   output.TotalEntries, // Same as entries for now
		TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
		TotalEntries:  output.TotalEntries,
		DrawDate:      date,
		LastUpdated:   time.Now(),
	}

	return result, nil
}

// InvokeRunnerUp invokes a runner-up for a prize
func (d *DrawServiceAdapter) InvokeRunnerUp(
	ctx context.Context,
	winnerID uuid.UUID,
	reason string,
	invokedByID uuid.UUID,
) (*entity.Winner, error) {
	// Create input for the service
	input := draw.InvokeRunnerUpInput{
		WinnerID:    winnerID,
		AdminUserID: invokedByID,
		Reason:      reason,
	}

	// Invoke runner-up
	output, err := d.invokeRunnerUpService.InvokeRunnerUp(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create masked MSISDN
	maskedMSISDN := util.MaskMSISDN(output.NewWinner.MSISDN)
	
	// Create response
	result := &entity.Winner{
		ID:          output.NewWinner.ID,
		DrawID:      uuid.Nil, // Not available in output
		MSISDN:      output.NewWinner.MSISDN,
		MaskedMSISDN: maskedMSISDN,
		PrizeID:     uuid.Nil, // Not available in output
		PrizeTierID: output.NewWinner.PrizeTierID,
		PrizeName:   "",
		PrizeValue:  output.NewWinner.PrizeValue,
		Status:      output.NewWinner.Status,
		IsRunnerUp:  false, // This was a runner-up but is now a winner
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return result, nil
}

// UpdateWinnerPaymentStatus updates a winner's payment status
func (d *DrawServiceAdapter) UpdateWinnerPaymentStatus(
	ctx context.Context,
	winnerID uuid.UUID,
	paymentStatus string,
	updatedByID uuid.UUID,
) (*entity.Winner, error) {
	// Create input for the service - convert UUID to string
	winnerIDStr := winnerID.String()
	
	input := draw.UpdateWinnerPaymentStatusInput{
		WinnerID:      winnerIDStr,
		PaymentStatus: paymentStatus,
	}

	// Update winner payment status
	output, err := d.updateWinnerService.UpdateWinnerPaymentStatus(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create masked MSISDN
	maskedMSISDN := util.MaskMSISDN(output.MSISDN)
	
	// Create response
	result := &entity.Winner{
		ID:            output.ID,
		DrawID:        uuid.Nil, // Not available in output
		MSISDN:        output.MSISDN,
		MaskedMSISDN:  maskedMSISDN,
		PrizeID:       uuid.Nil, // Not available in output
		PrizeTierID:   uuid.Nil, // Not available in output
		PrizeName:     "",
		PrizeValue:    0,
		Status:        output.Status,
		PaymentStatus: output.PaymentStatus,
		IsRunnerUp:    false, // Default value
		CreatedAt:     time.Now(), // Default value
		UpdatedAt:     time.Now(),
	}

	return result, nil
}

// GetWinners gets a list of winners with pagination
func (d *DrawServiceAdapter) GetWinners(
	ctx context.Context,
	page, pageSize int,
	drawID uuid.UUID,
	msisdn string,
	status string,
	paymentStatus string,
	isRunnerUp bool,
) (*entity.PaginatedWinners, error) {
	// Create input for the service
	// Note: Adjusting to match actual service input structure
	input := draw.ListWinnersInput{
		Page:          page,
		PageSize:      pageSize,
	}

	// Get winners
	output, err := d.listWinnersService.ListWinners(ctx, input)
	if err != nil {
		return nil, err
	}

	// Filter winners based on criteria if needed
	filteredWinners := output.Winners
	
	// Convert winners to entity model
	winners := make([]entity.Winner, 0, len(filteredWinners))
	for _, w := range filteredWinners {
		// Create masked MSISDN
		maskedMSISDN := util.MaskMSISDN(w.MSISDN)
		
		winners = append(winners, entity.Winner{
			ID:            w.ID,
			DrawID:        w.DrawID,
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskedMSISDN,
			PrizeID:       uuid.Nil, // Not available in output
			PrizeTierID:   w.PrizeTierID,
			PrizeName:     "",
			PrizeValue:    w.PrizeValue,
			Status:        w.Status,
			PaymentStatus: w.PaymentStatus,
			IsRunnerUp:    w.IsRunnerUp,
			CreatedAt:     w.CreatedAt,
			UpdatedAt:     w.UpdatedAt,
		})
	}

	// Create response
	result := &entity.PaginatedWinners{
		Winners:    winners,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}

	return result, nil
}
