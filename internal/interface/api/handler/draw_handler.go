package handler

import (
	"net/http"
	"strconv"
	"time"
	"fmt"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	drawApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// DrawHandler handles HTTP requests related to draws
type DrawHandler struct {
	getDrawsService                *drawApp.GetDrawsService
	getDrawByIDService             *drawApp.GetDrawByIDService
	executeDrawService             *drawApp.ExecuteDrawService
	getWinnersService              *drawApp.GetWinnersService
	getEligibilityStatsService     *drawApp.GetEligibilityStatsService
	invokeRunnerUpService          *drawApp.InvokeRunnerUpService
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	getDrawsService *drawApp.GetDrawsService,
	getDrawByIDService *drawApp.GetDrawByIDService,
	executeDrawService *drawApp.ExecuteDrawService,
	getWinnersService *drawApp.GetWinnersService,
	getEligibilityStatsService *drawApp.GetEligibilityStatsService,
	invokeRunnerUpService *drawApp.InvokeRunnerUpService,
	updateWinnerPaymentStatusService *drawApp.UpdateWinnerPaymentStatusService,
) *DrawHandler {
	return &DrawHandler{
		getDrawsService:                getDrawsService,
		getDrawByIDService:             getDrawByIDService,
		executeDrawService:             executeDrawService,
		getWinnersService:              getWinnersService,
		getEligibilityStatsService:     getEligibilityStatsService,
		invokeRunnerUpService:          invokeRunnerUpService,
		updateWinnerPaymentStatusService: updateWinnerPaymentStatusService,
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
	
	// Prepare input
	input := drawApp.GetDrawsInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	// Get draws
	output, err := h.getDrawsService.GetDraws(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draws: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	draws := make([]response.DrawResponse, 0, len(output.Draws))
	for _, d := range output.Draws {
		draws = append(draws, response.DrawResponse{
			ID:             d.ID.String(),
			Name:           d.Name,
			Description:    d.Description,
			DrawDate:       util.FormatTimeOrEmpty(d.DrawDate, time.RFC3339),
			Status:         d.Status,
			PrizeStructure: d.PrizeStructureName,
			CreatedAt:      util.FormatTimeOrEmpty(d.CreatedAt, time.RFC3339),
			CreatedBy:      d.CreatedBy.String(),
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
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeTierName,
			PrizeValue:    fmt.Sprintf("%.2f", w.PrizeValue),
			PaymentStatus: w.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
			PaymentRef:    w.PaymentRef,
			Status:        w.Status,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:             output.Draw.ID.String(),
			Name:           output.Draw.Name,
			Description:    output.Draw.Description,
			DrawDate:       util.FormatTimeOrEmpty(output.Draw.DrawDate, time.RFC3339),
			Status:         output.Draw.Status,
			PrizeStructure: output.Draw.PrizeStructureName,
			Winners:        winners,
			CreatedAt:      util.FormatTimeOrEmpty(output.Draw.CreatedAt, time.RFC3339),
			CreatedBy:      output.Draw.CreatedBy.String(),
		},
	})
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
	
	// Prepare input
	input := drawApp.ExecuteDrawInput{
		Name:            req.Name,
		Description:     req.Description,
		DrawDate:        req.DrawDate,
		PrizeStructureID: req.PrizeStructureID,
		CreatedBy:       userID.(uuid.UUID),
	}
	
	// Execute draw
	output, err := h.executeDrawService.ExecuteDraw(c.Request.Context(), input)
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
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeTierName,
			PrizeValue:    fmt.Sprintf("%.2f", w.PrizeValue),
			PaymentStatus: w.PaymentStatus,
			Status:        w.Status,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:             output.Draw.ID.String(),
			Name:           output.Draw.Name,
			Description:    output.Draw.Description,
			DrawDate:       util.FormatTimeOrEmpty(output.Draw.DrawDate, time.RFC3339),
			Status:         output.Draw.Status,
			PrizeStructure: output.Draw.PrizeStructureName,
			Winners:        winners,
			CreatedAt:      util.FormatTimeOrEmpty(output.Draw.CreatedAt, time.RFC3339),
			CreatedBy:      output.Draw.CreatedBy.String(),
		},
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
	
	// Parse filter parameters
	drawIDStr := c.DefaultQuery("drawId", "")
	var drawID *uuid.UUID
	if drawIDStr != "" {
		id, err := uuid.Parse(drawIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid draw ID format",
			})
			return
		}
		drawID = &id
	}
	
	msisdn := c.DefaultQuery("msisdn", "")
	paymentStatus := c.DefaultQuery("paymentStatus", "")
	
	// Prepare input
	input := drawApp.GetWinnersInput{
		Page:          page,
		PageSize:      pageSize,
		DrawID:        drawID,
		MSISDN:        msisdn,
		PaymentStatus: paymentStatus,
	}
	
	// Get winners
	output, err := h.getWinnersService.GetWinners(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get winners: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	winners := make([]response.WinnerResponse, 0, len(output.Winners))
	for _, w := range output.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeTierName,
			PrizeValue:    fmt.Sprintf("%.2f", w.PrizeValue),
			PaymentStatus: w.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
			PaymentRef:    w.PaymentRef,
			Status:        w.Status,
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
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// GetEligibilityStats handles GET /api/admin/draws/eligibility-stats
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse date parameter
	dateStr := c.DefaultQuery("date", "")
	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid date format. Expected YYYY-MM-DD",
			})
			return
		}
	} else {
		date = time.Now()
	}
	
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
			Date:          dateStr,
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
	
	// Prepare input
	input := drawApp.InvokeRunnerUpInput{
		WinnerID:    winnerID,
		InvokedBy:   userID.(uuid.UUID),
		Reason:      req.Reason,
	}
	
	// Invoke runner-up
	output, err := h.invokeRunnerUpService.InvokeRunnerUp(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to invoke runner-up: " + err.Error(),
		})
		return
	}
	
	// Prepare response with complete winner information
	originalWinner := response.WinnerResponse{
		ID:            output.OriginalWinner.ID.String(),
		DrawID:        output.OriginalWinner.DrawID.String(),
		MSISDN:        output.OriginalWinner.MSISDN,
		MaskedMSISDN:  maskMSISDN(output.OriginalWinner.MSISDN),
		PrizeTierID:   output.OriginalWinner.PrizeTierID.String(),
		PrizeName:     output.OriginalWinner.PrizeTierName,
		PrizeValue:    fmt.Sprintf("%.2f", output.OriginalWinner.PrizeValue),
		PaymentStatus: output.OriginalWinner.PaymentStatus,
		Status:        output.OriginalWinner.Status,
		IsRunnerUp:    output.OriginalWinner.IsRunnerUp,
		RunnerUpRank:  output.OriginalWinner.RunnerUpRank,
		CreatedAt:     util.FormatTimeOrEmpty(output.OriginalWinner.CreatedAt, time.RFC3339),
	}
	
	newWinner := response.WinnerResponse{
		ID:            output.NewWinner.ID.String(),
		DrawID:        output.NewWinner.DrawID.String(),
		MSISDN:        output.NewWinner.MSISDN,
		MaskedMSISDN:  maskMSISDN(output.NewWinner.MSISDN),
		PrizeTierID:   output.NewWinner.PrizeTierID.String(),
		PrizeName:     output.NewWinner.PrizeTierName,
		PrizeValue:    fmt.Sprintf("%.2f", output.NewWinner.PrizeValue),
		PaymentStatus: output.NewWinner.PaymentStatus,
		Status:        output.NewWinner.Status,
		IsRunnerUp:    output.NewWinner.IsRunnerUp,
		RunnerUpRank:  output.NewWinner.RunnerUpRank,
		CreatedAt:     util.FormatTimeOrEmpty(output.NewWinner.CreatedAt, time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.RunnerUpInvocationResult{
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
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Convert userID to UUID
	var updatedBy uuid.UUID
	switch id := userIDStr.(type) {
	case uuid.UUID:
		updatedBy = id
	case string:
		var err error
		updatedBy, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format in token",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID type in token",
		})
		return
	}
	
	// Prepare input
	input := drawApp.UpdateWinnerPaymentStatusInput{
		WinnerID:      winnerID.String(),
		PaymentStatus: req.PaymentStatus,
		PaymentNotes:  req.PaymentRef,
		Notes:         req.PaymentRef,
		UpdatedBy:     updatedBy,
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
	
	// Prepare response with complete winner information
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.WinnerResponse{
			ID:            output.ID.String(),
			DrawID:        output.DrawID.String(),
			MSISDN:        output.MSISDN,
			MaskedMSISDN:  maskMSISDN(output.MSISDN),
			PrizeTierID:   output.PrizeTierID.String(),
			PrizeName:     output.PrizeTierName,
			PrizeValue:    fmt.Sprintf("%.2f", output.PrizeValue),
			PaymentStatus: output.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(output.PaymentDate, time.RFC3339),
			PaymentRef:    output.PaymentRef,
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
