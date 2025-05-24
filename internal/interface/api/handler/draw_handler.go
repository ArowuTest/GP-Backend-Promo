package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/adapter"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	drawServiceAdapter *adapter.DrawServiceAdapter
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	drawServiceAdapter *adapter.DrawServiceAdapter,
) *DrawHandler {
	return &DrawHandler{
		drawServiceAdapter: drawServiceAdapter,
	}
}

// GetDraws handles GET /api/admin/draws
func (h *DrawHandler) GetDraws(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Get draws through adapter
	output, err := h.drawServiceAdapter.ListDraws(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draws: " + err.Error(),
		})
		return
	}

	// Prepare response with explicit type conversions at DTO boundary
	draws := make([]response.DrawResponse, 0, len(output.Draws))
	for _, d := range output.Draws {
		// Create a response that matches the frontend expectations
		draws = append(draws, response.DrawResponse{
			ID:             d.ID,
			Name:           d.Name,
			Description:    d.Description,
			DrawDate:       util.FormatTimeOrEmpty(d.DrawDate, "2006-01-02"),
			Status:         d.Status,
			PrizeStructure: d.PrizeStructureID.String(),
			CreatedAt:      util.FormatTimeOrEmpty(d.CreatedAt, time.RFC3339),
			CreatedBy:      d.ExecutedByAdminID.String(),
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    draws,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
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

	// Get draw through adapter
	output, err := h.drawServiceAdapter.GetDrawByID(c.Request.Context(), drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draw: " + err.Error(),
		})
		return
	}

	// Prepare response with explicit type conversions at DTO boundary
	winners := make([]response.WinnerResponse, 0)
	
	// Add winners if available in output
	if output != nil && len(output.Winners) > 0 {
		for _, w := range output.Winners {
			winners = append(winners, response.WinnerResponse{
				ID:            w.ID,
				DrawID:        w.DrawID.String(),
				MSISDN:        w.MSISDN,
				MaskedMSISDN:  maskMSISDN(w.MSISDN),
				PrizeTierID:   w.PrizeTierID.String(),
				PrizeName:     w.PrizeName,
				PrizeValue:    fmt.Sprintf("%.2f", w.PrizeValue),
				Status:        w.Status,
				PaymentStatus: w.PaymentStatus,
				IsRunnerUp:    w.IsRunnerUp,
				RunnerUpRank:  w.RunnerUpRank,
				CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
			})
		}
	}
		
	// Create a response that matches the frontend expectations
	drawResponse := response.DrawResponse{
		ID:             output.ID,
		Name:           output.Name,
		Description:    output.Description,
		DrawDate:       util.FormatTimeOrEmpty(output.DrawDate, "2006-01-02"),
		Status:         output.Status,
		PrizeStructure: output.PrizeStructureID.String(),
		Winners:        winners,
		CreatedAt:      util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
		CreatedBy:      output.ExecutedByAdminID.String(),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    drawResponse,
	})
}

// GetWinners handles GET /api/admin/winners
func (h *DrawHandler) GetWinners(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Parse filters if provided - not using these yet but keeping for future implementation
	_ = c.DefaultQuery("startDate", "")
	_ = c.DefaultQuery("endDate", "")

	// Get winners through adapter - using GetWinners instead of ListWinners to match adapter method
	output, err := h.drawServiceAdapter.GetWinners(c.Request.Context(), page, pageSize, uuid.Nil, "", "", "", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get winners: " + err.Error(),
		})
		return
	}

	// Prepare response with explicit type conversions at DTO boundary
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, w := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID,
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeName,
			PrizeValue:    fmt.Sprintf("%.2f", w.PrizeValue),
			Status:        w.Status,
			PaymentStatus: w.PaymentStatus,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    winners,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// ExecuteDraw handles POST /api/admin/draws
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req struct {
		DrawDate        string    `json:"drawDate" binding:"required"`
		PrizeStructureID uuid.UUID `json:"prizeStructureId" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	var executedBy uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		executedBy = id
	case string:
		var err error
		executedBy, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format in token",
				Details: "User ID must be a valid UUID",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID type in token",
			Details: "User ID must be a UUID or string",
		})
		return
	}

	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format, expected YYYY-MM-DD",
		})
		return
	}

	// Execute draw through adapter
	output, err := h.drawServiceAdapter.ExecuteDraw(c.Request.Context(), drawDate, req.PrizeStructureID, executedBy, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to execute draw: " + err.Error(),
		})
		return
	}

	// Create a response that matches the frontend expectations
	drawResponse := response.DrawResponse{
		ID:             output.ID,
		Name:           output.Name,
		Description:    output.Description,
		DrawDate:       req.DrawDate,
		Status:         output.Status,
		PrizeStructure: req.PrizeStructureID.String(),
		CreatedAt:      util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
		CreatedBy:      executedBy.String(),
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "Draw executed successfully",
		Data:    drawResponse,
	})
}

// GetEligibilityStats handles GET /api/admin/draws/eligibility-stats
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse draw date
	drawDateStr := c.DefaultQuery("drawDate", time.Now().Format("2006-01-02"))
	drawDate, err := time.Parse("2006-01-02", drawDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format, expected YYYY-MM-DD",
		})
		return
	}

	// Get eligibility stats through adapter
	output, err := h.drawServiceAdapter.GetEligibilityStats(c.Request.Context(), drawDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get eligibility stats: " + err.Error(),
		})
		return
	}

	// Prepare response that matches frontend expectations
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.EligibilityStatsResponse{
			Date:          drawDateStr,
			TotalEligible: output.TotalEligible,
			TotalEntries:  output.TotalEntries,
		},
	})
}

// InvokeRunnerUp handles POST /api/admin/winners/:id/invoke-runner-up
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	// Parse winner ID
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	var invokedBy uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		invokedBy = id
	case string:
		var err error
		invokedBy, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format in token",
				Details: "User ID must be a valid UUID",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID type in token",
			Details: "User ID must be a UUID or string",
		})
		return
	}

	// Parse request body
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Invoke runner-up through adapter
	newWinner, err := h.drawServiceAdapter.InvokeRunnerUp(c.Request.Context(), winnerID, req.Reason, invokedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
		})
		return
	}

	// Prepare response with new winner information
	newWinnerResponse := response.WinnerResponse{
		ID:            newWinner.ID,
		DrawID:        newWinner.DrawID.String(),
		MSISDN:        newWinner.MSISDN,
		MaskedMSISDN:  maskMSISDN(newWinner.MSISDN),
		PrizeTierID:   newWinner.PrizeTierID.String(),
		PrizeName:     newWinner.PrizeName,
		PrizeValue:    fmt.Sprintf("%.2f", newWinner.PrizeValue),
		PaymentStatus: newWinner.PaymentStatus,
		Status:        newWinner.Status,
		IsRunnerUp:    newWinner.IsRunnerUp,
		RunnerUpRank:  newWinner.RunnerUpRank,
		CreatedAt:     util.FormatTimeOrEmpty(newWinner.CreatedAt, time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.RunnerUpInvocationResult{
			Message:        "Runner-up successfully invoked",
			OriginalWinner: response.WinnerResponse{}, // Not available in current implementation
			NewWinner:      newWinnerResponse,
		},
	})
}

// UpdateWinnerPaymentStatus handles PUT /api/admin/winners/:id/payment-status
func (h *DrawHandler) UpdateWinnerPaymentStatus(c *gin.Context) {
	// Parse winner ID
	winnerIDStr := c.Param("id")
	if winnerIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Winner ID is required",
		})
		return
	}
	
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
		})
		return
	}

	var req struct {
		PaymentStatus string `json:"paymentStatus" binding:"required"`
		PaymentRef    string `json:"paymentRef"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	var updatedBy uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		updatedBy = id
	case string:
		var err error
		updatedBy, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format in token",
				Details: "User ID must be a valid UUID",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID type in token",
			Details: "User ID must be a UUID or string",
		})
		return
	}

	// Update winner payment status through adapter
	output, err := h.drawServiceAdapter.UpdateWinnerPaymentStatus(c.Request.Context(), winnerID, req.PaymentStatus, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update winner payment status: " + err.Error(),
		})
		return
	}

	// Prepare response with complete winner information
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.WinnerResponse{
			ID:            output.ID,
			DrawID:        output.DrawID.String(),
			MSISDN:        output.MSISDN,
			MaskedMSISDN:  maskMSISDN(output.MSISDN),
			PrizeTierID:   output.PrizeTierID.String(),
			PrizeName:     output.PrizeName,
			PrizeValue:    fmt.Sprintf("%.2f", output.PrizeValue),
			PaymentStatus: output.PaymentStatus,
			PaymentNotes:  output.PaymentNotes,
			Status:        output.Status,
			IsRunnerUp:    output.IsRunnerUp,
			RunnerUpRank:  output.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
			UpdatedAt:     util.FormatTimeOrEmpty(output.UpdatedAt, time.RFC3339),
		},
	})
}

// Helper function to mask MSISDN
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 6 {
		return msisdn
	}

	prefix := msisdn[:3]
	suffix := msisdn[len(msisdn)-3:]
	masked := prefix + "****" + suffix

	return masked
}
