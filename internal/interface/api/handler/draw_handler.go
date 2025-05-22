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
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	executeDrawService          *drawApp.ExecuteDrawService
	getDrawByIDService          *drawApp.GetDrawByIDService
	listDrawsService            *drawApp.ListDrawsService
	listWinnersService          *drawApp.ListWinnersService
	getEligibilityStatsService  *drawApp.GetEligibilityStatsService
	invokeRunnerUpService       *drawApp.InvokeRunnerUpService
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	executeDrawService *drawApp.ExecuteDrawService,
	getDrawByIDService *drawApp.GetDrawByIDService,
	listDrawsService *drawApp.ListDrawsService,
	listWinnersService *drawApp.ListWinnersService,
	getEligibilityStatsService *drawApp.GetEligibilityStatsService,
	invokeRunnerUpService *drawApp.InvokeRunnerUpService,
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService,
) *DrawHandler {
	return &DrawHandler{
		executeDrawService:      executeDrawService,
		getDrawByIDService:      getDrawByIDService,
		listDrawsService:        listDrawsService,
		listWinnersService:      listWinnersService,
		getEligibilityStatsService: getEligibilityStatsService,
		invokeRunnerUpService:   invokeRunnerUpService,
		updateWinnerPaymentStatusService: updateWinnerPaymentStatusService,
	}
}

// ExecuteDraw handles POST /api/admin/draws/execute
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req request.ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
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
		})
		return
	}
	
	// Parse draw date string to time.Time
	drawDate := util.ParseTimeOrZero(req.DrawDate, "2006-01-02")
	if drawDate.IsZero() {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format. Expected YYYY-MM-DD.",
		})
		return
	}
	
	// Prepare input
	input := drawApp.ExecuteDrawInput{
		DrawDate:         drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedByAdminID: userID.(uuid.UUID),
	}
	
	// Execute draw - removed context parameter to match service signature
	output, err := h.executeDrawService.ExecuteDraw(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to execute draw: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, w := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        output.DrawID.String(),
			MSISDN:        w.MSISDN,
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeName, // Map PrizeName to PrizeTierName
			PrizeValue:    w.PrizeValue, // PrizeValue is already a string in WinnerOutput
			Status:        "PendingNotification",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.DrawID.String(),
			DrawDate:             util.FormatTimeOrEmpty(output.DrawDate, "2006-01-02"),
			PrizeStructureID:     input.PrizeStructureID.String(),
			Status:               "Completed",
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
			ExecutedByAdminID:    input.ExecutedByAdminID.String(),
			Winners:              winners,
			CreatedAt:            time.Now().Format(time.RFC3339),
			UpdatedAt:            time.Now().Format(time.RFC3339),
		},
	})
}

// GetDrawByID handles GET /api/admin/draws/:id
func (h *DrawHandler) GetDrawByID(c *gin.Context) {
	// Parse draw ID
	drawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw ID format",
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
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, w := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        output.ID.String(),
			MSISDN:        w.MSISDN,
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeTierName, // Using PrizeTierName from domain entity
			PrizeValue:    w.PrizeValue,    // Using PrizeValue from domain entity
			Status:        w.Status,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.ID.String(),
			DrawDate:             util.FormatTimeOrEmpty(output.DrawDate, "2006-01-02"),
			PrizeStructureID:     output.PrizeStructureID.String(),
			Status:               output.Status,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
			ExecutedByAdminID:    output.ExecutedBy.String(),
			Winners:              winners,
			CreatedAt:            util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
			UpdatedAt:            util.FormatTimeOrEmpty(output.UpdatedAt, time.RFC3339),
		},
	})
}

// ListDraws handles GET /api/admin/draws
func (h *DrawHandler) ListDraws(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Prepare input
	input := drawApp.ListDrawsInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	// List draws
	output, err := h.listDrawsService.ListDraws(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list draws: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	draws := make([]response.DrawResponse, 0, len(output.Draws))
	for _, d := range output.Draws {
		winners := make([]response.WinnerResponse, 0, len(d.Winners))
		for _, w := range d.Winners {
			winners = append(winners, response.WinnerResponse{
				ID:            w.ID.String(),
				DrawID:        d.ID.String(),
				MSISDN:        w.MSISDN,
				PrizeTierID:   w.PrizeTierID.String(),
				PrizeTierName: w.PrizeTierName, // Using PrizeTierName from domain entity
				PrizeValue:    w.PrizeValue,    // Using PrizeValue from domain entity
				Status:        w.Status,
				IsRunnerUp:    w.IsRunnerUp,
				RunnerUpRank:  w.RunnerUpRank,
				CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
			})
		}
		
		draws = append(draws, response.DrawResponse{
			ID:                   d.ID.String(),
			DrawDate:             util.FormatTimeOrEmpty(d.DrawDate, "2006-01-02"),
			PrizeStructureID:     d.PrizeStructureID.String(),
			Status:               d.Status,
			TotalEligibleMSISDNs: d.TotalEligibleMSISDNs,
			TotalEntries:         d.TotalEntries,
			ExecutedByAdminID:    d.ExecutedBy.String(),
			Winners:              winners,
			CreatedAt:            util.FormatTimeOrEmpty(d.CreatedAt, time.RFC3339),
			UpdatedAt:            util.FormatTimeOrEmpty(d.UpdatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    draws,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
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
	
	// Parse date range parameters
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
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, w := range output.Winners {
		// Create response with available fields
		winnerResponse := response.WinnerResponse{
			ID:            w.ID.String(),
			MSISDN:        w.MSISDN,
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeTierName, // Using PrizeTierName from domain entity
			PrizeValue:    w.PrizeValue,    // Using PrizeValue from domain entity
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
			// Set default values for fields that might not exist
			DrawID:        "",
			Status:        "PendingNotification",
			PaymentStatus: "Pending",
			PaymentNotes:  "",
			PaidAt:        "",
		}
		
		winners = append(winners, winnerResponse)
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

// GetEligibilityStats handles GET /api/admin/draws/eligibility-stats
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse date parameter
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	
	// Prepare input
	input := drawApp.GetEligibilityStatsInput{
		Date: date,
	}
	
	// Get eligibility stats
	output, err := h.getEligibilityStatsService.GetEligibilityStats(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get eligibility stats: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.EligibilityStatsResponse{
			Date:                 output.Date,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
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
	
	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
		})
		return
	}
	
	// Prepare input - using the correct field names from the application layer
	input := drawApp.InvokeRunnerUpInput{
		WinnerID:    winnerID,
		AdminUserID: userID.(uuid.UUID),
		Reason:      req.Reason,
	}
	
	// Invoke runner-up - do not pass context as it's not in the method signature
	output, err := h.invokeRunnerUpService.InvokeRunnerUp(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
		})
		return
	}
	
	// Prepare response - using only the fields that exist in RunnerUpWinnerOutput
	originalWinner := response.WinnerResponse{
		ID:          output.OriginalWinner.ID.String(),
		MSISDN:      output.OriginalWinner.MSISDN,
		PrizeTierID: output.OriginalWinner.PrizeTierID.String(),
		Status:      output.OriginalWinner.Status,
		// Set default values for fields that don't exist in RunnerUpWinnerOutput
		DrawID:        "",
		PrizeTierName: "",
		PrizeValue:    "",
		IsRunnerUp:    false,
		RunnerUpRank:  0,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	
	newWinner := response.WinnerResponse{
		ID:          output.NewWinner.ID.String(),
		MSISDN:      output.NewWinner.MSISDN,
		PrizeTierID: output.NewWinner.PrizeTierID.String(),
		Status:      output.NewWinner.Status,
		// Set default values for fields that don't exist in RunnerUpWinnerOutput
		DrawID:        "",
		PrizeTierName: "",
		PrizeValue:    "",
		IsRunnerUp:    false,
		RunnerUpRank:  0,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.RunnerUpResponse{
			Message:        "Runner-up successfully invoked",
			OriginalWinner: originalWinner,
			NewWinner:      newWinner,
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
		})
		return
	}
	
	var req request.UpdateWinnerPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
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
			CreatedAt:     util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
		},
	})
}
