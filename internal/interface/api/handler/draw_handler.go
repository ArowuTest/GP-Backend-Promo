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
	// Get draws
	output, err := h.getDrawsService.GetDraws(c.Request.Context())
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
		// Convert winners
		winners := make([]response.WinnerResponse, 0, len(d.Winners))
		for _, w := range d.Winners {
			winners = append(winners, response.WinnerResponse{
				ID:            w.ID.String(),
				DrawID:        w.DrawID.String(),
				MSISDN:        w.MSISDN,
				MaskedMSISDN:  maskMSISDN(w.MSISDN),
				PrizeTierID:   w.PrizeTierID.String(),
				PrizeName:     w.PrizeName,
				PrizeValue:    w.PrizeValue,
				PaymentStatus: w.PaymentStatus,
				PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
				PaymentRef:    w.PaymentRef,
				IsRunnerUp:    w.IsRunnerUp,
				RunnerUpRank:  w.RunnerUpRank,
				Status:        w.Status,
				InvokedAt:     util.FormatTimeOrEmpty(w.InvokedAt, time.RFC3339),
				CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
			})
		}
		
		draws = append(draws, response.DrawResponse{
			ID:             d.ID.String(),
			Name:           d.Name,
			Description:    d.Description,
			DrawDate:       util.FormatTimeOrEmpty(d.DrawDate, "2006-01-02"),
			Status:         d.Status,
			PrizeStructure: d.PrizeStructureID.String(),
			Winners:        winners,
			CreatedAt:      util.FormatTimeOrEmpty(d.CreatedAt, time.RFC3339),
			CreatedBy:      d.CreatedBy.String(),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    draws,
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
	
	// Get draw by ID
	output, err := h.getDrawByIDService.GetDrawByID(c.Request.Context(), drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get draw: " + err.Error(),
		})
		return
	}
	
	// Convert winners
	winners := make([]response.WinnerResponse, 0, len(output.Draw.Winners))
	for _, w := range output.Draw.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeName,
			PrizeValue:    w.PrizeValue,
			PaymentStatus: w.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
			PaymentRef:    w.PaymentRef,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			Status:        w.Status,
			InvokedAt:     util.FormatTimeOrEmpty(w.InvokedAt, time.RFC3339),
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DrawResponse{
			ID:             output.Draw.ID.String(),
			Name:           output.Draw.Name,
			Description:    output.Draw.Description,
			DrawDate:       util.FormatTimeOrEmpty(output.Draw.DrawDate, "2006-01-02"),
			Status:         output.Draw.Status,
			PrizeStructure: output.Draw.PrizeStructureID.String(),
			Winners:        winners,
			CreatedAt:      util.FormatTimeOrEmpty(output.Draw.CreatedAt, time.RFC3339),
			CreatedBy:      output.Draw.CreatedBy.String(),
		},
	})
}

// GetWinners handles GET /api/admin/winners
func (h *DrawHandler) GetWinners(c *gin.Context) {
	// Get winners
	output, err := h.getWinnersService.GetWinners(c.Request.Context())
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
			ID:            w.ID.String(),
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeName,
			PrizeValue:    w.PrizeValue,
			PaymentStatus: w.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
			PaymentRef:    w.PaymentRef,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			Status:        w.Status,
			InvokedAt:     util.FormatTimeOrEmpty(w.InvokedAt, time.RFC3339),
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    winners,
	})
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
	
	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format, expected YYYY-MM-DD",
		})
		return
	}
	
	// Prepare input
	input := drawApp.ExecuteDrawInput{
		DrawDate:        drawDate,
		PrizeStructureID: prizeStructureID,
		AdminUserID:     userID.(uuid.UUID),
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
	
	// Convert winners
	winners := make([]response.WinnerResponse, 0, len(output.Draw.Winners))
	for _, w := range output.Draw.Winners {
		winners = append(winners, response.WinnerResponse{
			ID:            w.ID.String(),
			DrawID:        w.DrawID.String(),
			MSISDN:        w.MSISDN,
			MaskedMSISDN:  maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeName:     w.PrizeName,
			PrizeValue:    w.PrizeValue,
			PaymentStatus: w.PaymentStatus,
			PaymentDate:   util.FormatTimeOrEmpty(w.PaymentDate, time.RFC3339),
			PaymentRef:    w.PaymentRef,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
			Status:        w.Status,
			InvokedAt:     util.FormatTimeOrEmpty(w.InvokedAt, time.RFC3339),
			CreatedAt:     util.FormatTimeOrEmpty(w.CreatedAt, time.RFC3339),
		})
	}
	
	// Prepare response
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "Draw executed successfully",
		Data: response.DrawResponse{
			ID:             output.Draw.ID.String(),
			Name:           output.Draw.Name,
			Description:    output.Draw.Description,
			DrawDate:       util.FormatTimeOrEmpty(output.Draw.DrawDate, "2006-01-02"),
			Status:         output.Draw.Status,
			PrizeStructure: output.Draw.PrizeStructureID.String(),
			Winners:        winners,
			CreatedAt:      util.FormatTimeOrEmpty(output.Draw.CreatedAt, time.RFC3339),
			CreatedBy:      output.Draw.CreatedBy.String(),
		},
	})
}

// GetEligibilityStats handles GET /api/admin/draws/eligibility-stats
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse date
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid date format, expected YYYY-MM-DD",
		})
		return
	}
	
	// Get eligibility stats
	output, err := h.getEligibilityStatsService.GetEligibilityStats(c.Request.Context(), date)
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
			TotalEligible: output.TotalEligibleMSISDNs,
			TotalEntries:  output.TotalEntries,
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
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Convert userID to UUID
	var adminUserID uuid.UUID
	switch id := userIDStr.(type) {
	case uuid.UUID:
		adminUserID = id
	case string:
		var err error
		adminUserID, err = uuid.Parse(id)
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
	
	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID format",
		})
		return
	}
	
	// Prepare input
	input := drawApp.InvokeRunnerUpInput{
		WinnerID:    winnerID,
		AdminUserID: adminUserID,
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
		PrizeName:     output.OriginalWinner.PrizeName,
		PrizeValue:    output.OriginalWinner.PrizeValue,
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
		PrizeName:     output.NewWinner.PrizeName,
		PrizeValue:    output.NewWinner.PrizeValue,
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
		PaymentNotes:  req.Notes,
		Notes:         req.Notes,
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
			PrizeName:     output.PrizeName,
			PrizeValue:    output.PrizeValue,
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
