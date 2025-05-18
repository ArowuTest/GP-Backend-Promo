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
	executeDraw          *drawApp.ExecuteDrawService
	getDrawDetails       *drawApp.GetDrawDetailsService
	listDraws            *drawApp.ListDrawsService
	getEligibilityStats  *drawApp.GetEligibilityStatsService
	invokeRunnerUp       *drawApp.InvokeRunnerUpService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	executeDraw *drawApp.ExecuteDrawService,
	getDrawDetails *drawApp.GetDrawDetailsService,
	listDraws *drawApp.ListDrawsService,
	getEligibilityStats *drawApp.GetEligibilityStatsService,
	invokeRunnerUp *drawApp.InvokeRunnerUpService,
) *DrawHandler {
	return &DrawHandler{
		executeDraw:         executeDraw,
		getDrawDetails:      getDrawDetails,
		listDraws:           listDraws,
		getEligibilityStats: getEligibilityStats,
		invokeRunnerUp:      invokeRunnerUp,
	}
}

// ExecuteDraw handles POST /api/admin/draws
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
	
	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format",
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
	
	// Prepare input
	input := drawApp.ExecuteDrawInput{
		DrawDate:         drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedByAdminID: userID.(uuid.UUID),
	}
	
	// Execute draw
	output, err := h.executeDraw.ExecuteDraw(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to execute draw: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, winner := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            winner.ID.String(),
			MSISDN:        maskMSISDN(winner.MSISDN),
			PrizeTierID:   winner.PrizeTierID.String(),
			PrizeTierName: winner.PrizeValue,
			PrizeValue:    winner.PrizeValue,
			Status:        "PendingNotification",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.DrawID.String(),
			DrawDate:             output.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     input.PrizeStructureID.String(),
			Status:               "Completed",
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:         output.TotalEntries,
			ExecutedByAdminID:    input.ExecutedByAdminID.String(),
			CreatedAt:            time.Now().Format("2006-01-02 15:04:05"),
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
	input := drawApp.GetDrawDetailsInput{
		DrawID: drawID,
	}
	
	// Get draw details
	output, err := h.getDrawDetails.GetDrawDetails(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draw details: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, winner := range output.Winners {
		prizeTierName := ""
		prizeValue := ""
		
		winners = append(winners, response.WinnerResponse{
			ID:            winner.ID.String(),
			MSISDN:        maskMSISDN(winner.MSISDN),
			PrizeTierID:   winner.PrizeTierID.String(),
			PrizeTierName: prizeTierName,
			PrizeValue:    prizeValue,
			IsRunnerUp:    winner.IsRunnerUp,
			Status:        winner.Status,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                   output.Draw.ID.String(),
			DrawDate:             output.Draw.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     output.Draw.PrizeStructureID.String(),
			Status:               output.Draw.Status,
			Winners:              winners,
			ExecutedByAdminID:    output.Draw.ExecutedByAdminID.String(),
			CreatedAt:            output.Draw.CreatedAt.Format("2006-01-02 15:04:05"),
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
	output, err := h.listDraws.ListDraws(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list draws: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	draws := make([]response.DrawResponse, 0, len(output.Draws))
	for _, draw := range output.Draws {
		draws = append(draws, response.DrawResponse{
			ID:                   draw.ID.String(),
			DrawDate:             draw.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     draw.PrizeStructureID.String(),
			Status:               draw.Status,
			TotalEligibleMSISDNs: draw.TotalEligibleMSISDNs,
			TotalEntries:         draw.TotalEntries,
			ExecutedByAdminID:    draw.ExecutedByAdminID.String(),
			CreatedAt:            draw.CreatedAt.Format("2006-01-02 15:04:05"),
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

// GetEligibilityStats handles GET /api/admin/draws/eligibility-stats
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse date parameter
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	
	// Prepare input
	input := drawApp.GetEligibilityStatsInput{
		Date: date,
	}
	
	// Get eligibility stats
	output, err := h.getEligibilityStats.GetEligibilityStats(c.Request.Context(), input)
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
	
	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
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
		WinnerID:    winnerID,
		AdminUserID: userID.(uuid.UUID),
		Reason:      req.Reason,
	}
	
	// Invoke runner-up
	output, err := h.invokeRunnerUp.InvokeRunnerUp(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.RunnerUpResponse{
			Message: "Runner-up successfully invoked",
			OriginalWinner: response.WinnerResponse{
				ID:          output.OriginalWinner.ID.String(),
				MSISDN:      maskMSISDN(output.OriginalWinner.MSISDN),
				PrizeTierID: output.OriginalWinner.PrizeTierID.String(),
				Status:      output.OriginalWinner.Status,
			},
			NewWinner: response.WinnerResponse{
				ID:          output.NewWinner.ID.String(),
				MSISDN:      maskMSISDN(output.NewWinner.MSISDN),
				PrizeTierID: output.NewWinner.PrizeTierID.String(),
				Status:      output.NewWinner.Status,
			},
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
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a mock response with pagination
	
	// Mock winners
	winners := []response.WinnerResponse{
		{
			ID:            uuid.New().String(),
			MSISDN:        "234*****789", // Masked for privacy
			PrizeTierID:   uuid.New().String(),
			PrizeTierName: "Cash Prize",
			PrizeValue:    "N100,000",
			Status:        "Pending",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
		},
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    winners,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  len(winners),
			TotalPages: 1,
			TotalItems: int64(len(winners)),
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
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a success response
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.WinnerResponse{
			ID:            winnerID.String(),
			Status:        req.PaymentStatus,
			PaymentStatus: req.PaymentStatus,
			PaymentNotes:  req.Notes,
		},
	})
}

// Helper function to mask MSISDN (show only first 3 and last 3 digits)
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 6 {
		return msisdn
	}
	
	first3 := msisdn[:3]
	last3 := msisdn[len(msisdn)-3:]
	masked := first3 + "****" + last3
	
	return masked
}
