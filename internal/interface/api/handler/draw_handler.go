package handler

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	drawApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	getDrawByIDService *drawApp.GetDrawByIDService
	listWinnersService *drawApp.ListWinnersService
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService
	executeDrawService *drawApp.ExecuteDrawService
	invokeRunnerUpService *drawApp.InvokeRunnerUpService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	getDrawByIDService *drawApp.GetDrawByIDService,
	listWinnersService *drawApp.ListWinnersService,
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService,
	executeDrawService *drawApp.ExecuteDrawService,
	invokeRunnerUpService *drawApp.InvokeRunnerUpService,
) *DrawHandler {
	return &DrawHandler{
		getDrawByIDService: getDrawByIDService,
		listWinnersService: listWinnersService,
		updateWinnerPaymentStatusService: updateWinnerPaymentStatusService,
		executeDrawService: executeDrawService,
		invokeRunnerUpService: invokeRunnerUpService,
	}
}

// ExecuteDraw handles POST /api/admin/draws
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req request.ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check all required fields and formats",
		})
		return
	}
	
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
			Details: "The provided prize structure ID is not in the correct UUID format",
		})
		return
	}
	
	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format",
			Details: "The draw date must be in YYYY-MM-DD format",
		})
		return
	}
	
	// Prepare input
	input := drawApp.ExecuteDrawInput{
		DrawDate:         drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedBy:       userID.(uuid.UUID),
	}
	
	// Execute draw
	output, err := h.executeDrawService.ExecuteDraw(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to execute draw: " + err.Error(),
			Details: "An error occurred while executing the draw. Please try again later.",
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, winner := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            winner.ID.String(),
			DrawID:        winner.DrawID.String(),
			MSISDN:        winner.MSISDN,
			PrizeTierID:   winner.PrizeTierID.String(),
			PrizeTierName: winner.PrizeTierName,
			PrizeValue:    winner.PrizeValue,
			Status:        winner.Status,
			PaymentStatus: winner.PaymentStatus,
			IsRunnerUp:    winner.IsRunnerUp,
			RunnerUpRank:  winner.RunnerUpRank,
			CreatedAt:     winner.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.ID.String(),
			DrawDate:             output.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     output.PrizeStructureID.String(),
			Status:               output.Status,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
			ExecutedByAdminID:    output.ExecutedBy.String(),
			Winners:              winners,
			CreatedAt:            output.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:            output.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// GetDraw handles GET /api/admin/draws/:id
func (h *DrawHandler) GetDraw(c *gin.Context) {
	// Parse draw ID
	drawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw ID format",
			Details: "The provided ID is not in the correct UUID format",
		})
		return
	}
	
	// Prepare input
	input := drawApp.GetDrawByIDInput{
		ID: drawID,
	}
	
	// Get draw
	output, err := h.getDrawByIDService.GetDrawByID(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draw: " + err.Error(),
			Details: "An error occurred while retrieving the draw. Please try again later.",
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, winner := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            winner.ID.String(),
			DrawID:        winner.DrawID.String(),
			MSISDN:        winner.MSISDN,
			PrizeTierID:   winner.PrizeTierID.String(),
			PrizeTierName: winner.PrizeTierName,
			PrizeValue:    winner.PrizeValue,
			Status:        winner.Status,
			PaymentStatus: winner.PaymentStatus,
			IsRunnerUp:    winner.IsRunnerUp,
			RunnerUpRank:  winner.RunnerUpRank,
			CreatedAt:     winner.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.ID.String(),
			DrawDate:             output.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     output.PrizeStructureID.String(),
			Status:               output.Status,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
			ExecutedByAdminID:    output.ExecutedBy.String(),
			Winners:              winners,
			CreatedAt:            output.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:            output.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// ListWinners handles GET /api/admin/winners
func (h *DrawHandler) ListWinners(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Parse date range
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	
	// Prepare input
	input := drawApp.ListWinnersInput{
		Page:      page,
		PageSize:  pageSize,
		StartDate: startDate,
		EndDate:   endDate,
	}
	
	// List winners
	output, err := h.listWinnersService.ListWinners(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list winners: " + err.Error(),
			Details: "An error occurred while retrieving winners. Please try again later.",
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, winner := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            winner.ID.String(),
			DrawID:        winner.DrawID.String(),
			MSISDN:        winner.MSISDN,
			PrizeTierID:   winner.PrizeTierID.String(),
			PrizeTierName: winner.PrizeTierName,
			PrizeValue:    winner.PrizeValue,
			Status:        winner.Status,
			PaymentStatus: winner.PaymentStatus,
			PaymentNotes:  winner.PaymentNotes,
			PaidAt:        winner.PaidAt,
			IsRunnerUp:    winner.IsRunnerUp,
			RunnerUpRank:  winner.RunnerUpRank,
			CreatedAt:     winner.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    winners,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// UpdateWinnerPaymentStatus handles PUT /api/admin/winners/:id/payment-status
func (h *DrawHandler) UpdateWinnerPaymentStatus(c *gin.Context) {
	// Parse winner ID
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
			Details: "The provided ID is not in the correct UUID format",
		})
		return
	}
	
	var req request.UpdateWinnerPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check all required fields and formats",
		})
		return
	}
	
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Prepare input
	input := drawApp.UpdateWinnerPaymentStatusInput{
		WinnerID:      winnerID,
		PaymentStatus: req.PaymentStatus,
		Notes:         req.Notes,
		UpdatedBy:     userID.(uuid.UUID),
	}
	
	// Update winner payment status
	output, err := h.updateWinnerPaymentStatusService.UpdateWinnerPaymentStatus(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update winner payment status: " + err.Error(),
			Details: "An error occurred while updating the payment status. Please try again later.",
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.WinnerResponse{
			ID:            output.ID.String(),
			DrawID:        output.DrawID.String(),
			MSISDN:        output.MSISDN,
			PrizeTierID:   output.PrizeTierID.String(),
			PrizeTierName: output.PrizeTierName,
			PrizeValue:    output.PrizeValue,
			Status:        output.Status,
			PaymentStatus: output.PaymentStatus,
			PaymentNotes:  output.PaymentNotes,
			PaidAt:        output.PaidAt,
			IsRunnerUp:    output.IsRunnerUp,
			RunnerUpRank:  output.RunnerUpRank,
			CreatedAt:     output.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     output.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// InvokeRunnerUp handles POST /api/admin/draws/invoke-runner-up
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	var req request.InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check all required fields and formats",
		})
		return
	}
	
	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
			Details: "The provided winner ID is not in the correct UUID format",
		})
		return
	}
	
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Prepare input
	input := drawApp.InvokeRunnerUpInput{
		WinnerID:  winnerID,
		Reason:    req.Reason,
		InvokedBy: userID.(uuid.UUID),
	}
	
	// Invoke runner-up
	output, err := h.invokeRunnerUpService.InvokeRunnerUp(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
			Details: "An error occurred while invoking the runner-up. Please try again later.",
		})
		return
	}
	
	// Prepare response
	originalWinner := response.WinnerResponse{
		ID:            output.OriginalWinner.ID.String(),
		DrawID:        output.OriginalWinner.DrawID.String(),
		MSISDN:        output.OriginalWinner.MSISDN,
		PrizeTierID:   output.OriginalWinner.PrizeTierID.String(),
		PrizeTierName: output.OriginalWinner.PrizeTierName,
		PrizeValue:    output.OriginalWinner.PrizeValue,
		Status:        output.OriginalWinner.Status,
		PaymentStatus: output.OriginalWinner.PaymentStatus,
		IsRunnerUp:    output.OriginalWinner.IsRunnerUp,
		RunnerUpRank:  output.OriginalWinner.RunnerUpRank,
		CreatedAt:     output.OriginalWinner.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	
	newWinner := response.WinnerResponse{
		ID:            output.NewWinner.ID.String(),
		DrawID:        output.NewWinner.DrawID.String(),
		MSISDN:        output.NewWinner.MSISDN,
		PrizeTierID:   output.NewWinner.PrizeTierID.String(),
		PrizeTierName: output.NewWinner.PrizeTierName,
		PrizeValue:    output.NewWinner.PrizeValue,
		Status:        output.NewWinner.Status,
		PaymentStatus: output.NewWinner.PaymentStatus,
		IsRunnerUp:    output.NewWinner.IsRunnerUp,
		RunnerUpRank:  output.NewWinner.RunnerUpRank,
		CreatedAt:     output.NewWinner.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.RunnerUpResponse{
			Message:         "Runner-up successfully invoked",
			OriginalWinner:  originalWinner,
			NewWinner:       newWinner,
		},
	})
}
