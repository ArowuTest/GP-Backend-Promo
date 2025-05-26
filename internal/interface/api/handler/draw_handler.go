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
	executeDrawService    *drawApp.ExecuteDrawService
	getDrawService        *drawApp.GetDrawByIDService
	listDrawsService      *drawApp.ListDrawsService
	invokeRunnerUpService *drawApp.InvokeRunnerUpService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	executeDrawService *drawApp.ExecuteDrawService,
	getDrawService *drawApp.GetDrawByIDService,
	listDrawsService *drawApp.ListDrawsService,
	invokeRunnerUpService *drawApp.InvokeRunnerUpService,
) *DrawHandler {
	return &DrawHandler{
		executeDrawService:    executeDrawService,
		getDrawService:        getDrawService,
		listDrawsService:      listDrawsService,
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
		})
		return
	}

	// Get user ID from context
	adminIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Prepare input
	input := drawApp.ExecuteDrawInput{
		DrawDate:         req.DrawDate,
		PrizeStructureID: req.PrizeStructureID,
		ExecutedByAdminID: adminIDValue.(uuid.UUID),
	}

	// Execute draw
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
			ID:           w.ID.String(),
			MSISDN:       w.MSISDN,
			MaskedMSISDN: maskMSISDN(w.MSISDN),
			PrizeTierID:  w.PrizeTierID.String(),
			PrizeTierName: w.PrizeName,
			PrizeValue:   strconv.FormatFloat(w.PrizeValue, 'f', 2, 64),
		})
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                  output.DrawID.String(),
			DrawDate:            output.DrawDate.Format("2006-01-02"),
			PrizeStructureID:    input.PrizeStructureID.String(),
			Status:              "Completed",
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:        output.TotalEntries,
			ExecutedByAdminID:   input.ExecutedByAdminID.String(),
			Winners:             winners,
			CreatedAt:           time.Now().Format("2006-01-02 15:04:05"),
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
		})
		return
	}

	// Prepare input
	input := drawApp.GetDrawByIDInput{
		ID: drawID,
	}

	// Get draw
	output, err := h.getDrawService.GetDrawByID(c.Request.Context(), input)
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
		// Use domain Winner struct fields, not WinnerOutput
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeTierName,
			PrizeValue:    strconv.FormatFloat(w.PrizeValue, 'f', 2, 64),
			Status:        w.Status,
			PaymentStatus: w.PaymentStatus,
			PaymentNotes:  w.PaymentNotes,
			PaidAt:        formatTimePtr(w.PaidAt),
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     w.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     formatTimePtr(&w.UpdatedAt),
		})
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:                  output.ID.String(),
			DrawDate:            output.DrawDate.Format("2006-01-02"),
			PrizeStructureID:    output.PrizeStructureID.String(),
			Status:              output.Status,
			TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
			TotalEntries:        output.TotalEntries,
			ExecutedByAdminID:   output.ExecutedBy.String(),
			Winners:             winners,
			CreatedAt:           output.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:           output.UpdatedAt.Format("2006-01-02 15:04:05"),
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
				DrawID:        w.DrawID.String(),
				MSISDN:        w.MSISDN,
				MaskedMSISDN:  maskMSISDN(w.MSISDN),
				PrizeTierID:   w.PrizeTierID.String(),
				PrizeTierName: w.PrizeTierName,
				PrizeValue:    strconv.FormatFloat(w.PrizeValue, 'f', 2, 64),
				Status:        w.Status,
				IsRunnerUp:    w.IsRunnerUp,
				RunnerUpRank:  w.RunnerUpRank,
				CreatedAt:     w.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		draws = append(draws, response.DrawResponse{
			ID:                  d.ID.String(),
			DrawDate:            d.DrawDate.Format("2006-01-02"),
			PrizeStructureID:    d.PrizeStructureID.String(),
			Status:              d.Status,
			TotalEligibleMSISDNs: d.TotalEligibleMSISDNs,
			TotalEntries:        d.TotalEntries,
			ExecutedByAdminID:   d.ExecutedBy.String(),
			Winners:             winners,
			CreatedAt:           d.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:           d.UpdatedAt.Format("2006-01-02 15:04:05"),
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

	var req request.InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	adminIDValue, exists := c.Get("userID")
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
		Reason:      req.Reason,
		AdminUserID: adminIDValue.(uuid.UUID),
	}

	// Invoke runner-up
	output, err := h.invokeRunnerUpService.InvokeRunnerUp(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
		})
		return
	}

	// Prepare response - ONLY using fields that exist in RunnerUpWinnerOutput
	originalWinner := response.WinnerResponse{
		ID:           output.OriginalWinner.ID.String(),
		MSISDN:       output.OriginalWinner.MSISDN,
		MaskedMSISDN: maskMSISDN(output.OriginalWinner.MSISDN),
		PrizeTierID:  output.OriginalWinner.PrizeTierID.String(),
		PrizeValue:   strconv.FormatFloat(output.OriginalWinner.PrizeValue, 'f', 2, 64),
		Status:       output.OriginalWinner.Status,
	}

	runnerUpWinner := response.WinnerResponse{
		ID:           output.RunnerUpWinner.ID.String(),
		MSISDN:       output.RunnerUpWinner.MSISDN,
		MaskedMSISDN: maskMSISDN(output.RunnerUpWinner.MSISDN),
		PrizeTierID:  output.RunnerUpWinner.PrizeTierID.String(),
		PrizeValue:   strconv.FormatFloat(output.RunnerUpWinner.PrizeValue, 'f', 2, 64),
		Status:       output.RunnerUpWinner.Status,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.InvokeRunnerUpResponse{
			OriginalWinner: originalWinner,
			RunnerUpWinner: runnerUpWinner,
			Reason:         req.Reason,
			InvokedBy:      adminIDValue.(uuid.UUID).String(),
			InvokedAt:      time.Now().Format("2006-01-02 15:04:05"),
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
	adminIDValue, exists := c.Get("userID")
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
		PaymentNotes:  req.PaymentNotes,
		UpdatedBy:     adminIDValue.(uuid.UUID),
	}

	// Update payment status
	err = h.executeDrawService.UpdateWinnerPaymentStatus(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update payment status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Payment status updated successfully",
		},
	})
}

// Helper function to mask MSISDN
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 4 {
		return msisdn
	}
	return "xxxx" + msisdn[len(msisdn)-4:]
}

// Helper function to format time pointer
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
