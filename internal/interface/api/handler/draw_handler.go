package interface

import (
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw/entity"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// DrawHandler handles HTTP requests related to draws
type DrawHandler struct {
	executeDraw       *draw.ExecuteDrawUseCase
	getDrawByID       *draw.GetDrawByIDUseCase
	listDraws         *draw.ListDrawsUseCase
	getEligibilityStats *draw.GetEligibilityStatsUseCase
	invokeRunnerUp    *draw.InvokeRunnerUpUseCase
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(
	executeDraw *draw.ExecuteDrawUseCase,
	getDrawByID *draw.GetDrawByIDUseCase,
	listDraws *draw.ListDrawsUseCase,
	getEligibilityStats *draw.GetEligibilityStatsUseCase,
	invokeRunnerUp *draw.InvokeRunnerUpUseCase,
) *DrawHandler {
	return &DrawHandler{
		executeDraw:       executeDraw,
		getDrawByID:       getDrawByID,
		listDraws:         listDraws,
		getEligibilityStats: getEligibilityStats,
		invokeRunnerUp:    invokeRunnerUp,
	}
}

// ExecuteDraw handles the request to execute a draw
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req request.ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Parse admin ID from JWT token
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User ID not found in token",
		})
		return
	}
	
	adminID, ok := adminIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Internal server error",
			Details: "Failed to parse user ID",
		})
		return
	}
	
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Internal server error",
			Details: "Invalid user ID format",
		})
		return
	}

	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw date format",
			Details: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	// Execute draw
	input := draw.ExecuteDrawInput{
		DrawDate:        drawDate,
		PrizeStructureID: prizeStructureID,
		AdminID:         adminUUID,
	}

	output, err := h.executeDraw.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.DrawError:
			drawErr := err.(*entity.DrawError)
			switch drawErr.Code() {
			case entity.ErrDrawAlreadyExists:
				statusCode = http.StatusConflict
				errorMessage = "A draw already exists for this date"
			case entity.ErrNoPrizeStructureActive:
				statusCode = http.StatusBadRequest
				errorMessage = "No active prize structure found for the given date"
			case entity.ErrNoEligibleParticipants:
				statusCode = http.StatusBadRequest
				errorMessage = "No eligible participants found for the draw"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to execute draw"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to execute draw"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert winners to response format
	winners := make([]response.WinnerResponse, 0, len(output.Draw.Winners))
	for _, w := range output.Draw.Winners {
		winner := response.WinnerResponse{
			ID:            w.ID.String(),
			MSISDN:        maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeTierName,
			PrizeValue:    w.PrizeValue,
			Status:        w.Status,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
		}
		winners = append(winners, winner)
	}

	// Prepare response
	resp := response.DrawResponse{
		ID:                   output.Draw.ID.String(),
		DrawDate:             output.Draw.DrawDate.Format("2006-01-02"),
		PrizeStructureID:     output.Draw.PrizeStructureID.String(),
		Status:               output.Draw.Status,
		TotalEligibleMSISDNs: output.Draw.TotalEligibleMSISDNs,
		TotalEntries:         output.Draw.TotalEntries,
		ExecutedByAdminID:    output.Draw.ExecutedByAdminID.String(),
		Winners:              winners,
		CreatedAt:            output.Draw.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// GetDrawByID handles the request to get a draw by ID
func (h *DrawHandler) GetDrawByID(c *gin.Context) {
	// Parse draw ID from URL
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid draw ID",
			Details: err.Error(),
		})
		return
	}

	// Get draw
	input := draw.GetDrawByIDInput{
		DrawID: drawID,
	}

	output, err := h.getDrawByID.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.DrawError:
			drawErr := err.(*entity.DrawError)
			if drawErr.Code() == entity.ErrDrawNotFound {
				statusCode = http.StatusNotFound
				errorMessage = "Draw not found"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to get draw"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to get draw"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert winners to response format
	winners := make([]response.WinnerResponse, 0, len(output.Draw.Winners))
	for _, w := range output.Draw.Winners {
		winner := response.WinnerResponse{
			ID:            w.ID.String(),
			MSISDN:        maskMSISDN(w.MSISDN),
			PrizeTierID:   w.PrizeTierID.String(),
			PrizeTierName: w.PrizeTierName,
			PrizeValue:    w.PrizeValue,
			Status:        w.Status,
			IsRunnerUp:    w.IsRunnerUp,
			RunnerUpRank:  w.RunnerUpRank,
		}
		winners = append(winners, winner)
	}

	// Prepare response
	resp := response.DrawResponse{
		ID:                   output.Draw.ID.String(),
		DrawDate:             output.Draw.DrawDate.Format("2006-01-02"),
		PrizeStructureID:     output.Draw.PrizeStructureID.String(),
		Status:               output.Draw.Status,
		TotalEligibleMSISDNs: output.Draw.TotalEligibleMSISDNs,
		TotalEntries:         output.Draw.TotalEntries,
		ExecutedByAdminID:    output.Draw.ExecutedByAdminID.String(),
		Winners:              winners,
		CreatedAt:            output.Draw.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// ListDraws handles the request to list draws
func (h *DrawHandler) ListDraws(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// List draws
	input := draw.ListDrawsInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := h.listDraws.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list draws",
			Details: err.Error(),
		})
		return
	}

	// Convert draws to response format
	draws := make([]response.DrawResponse, 0, len(output.Draws))
	for _, d := range output.Draws {
		// Convert winners to response format
		winners := make([]response.WinnerResponse, 0, len(d.Winners))
		for _, w := range d.Winners {
			winner := response.WinnerResponse{
				ID:            w.ID.String(),
				MSISDN:        maskMSISDN(w.MSISDN),
				PrizeTierID:   w.PrizeTierID.String(),
				PrizeTierName: w.PrizeTierName,
				PrizeValue:    w.PrizeValue,
				Status:        w.Status,
				IsRunnerUp:    w.IsRunnerUp,
				RunnerUpRank:  w.RunnerUpRank,
			}
			winners = append(winners, winner)
		}

		draw := response.DrawResponse{
			ID:                   d.ID.String(),
			DrawDate:             d.DrawDate.Format("2006-01-02"),
			PrizeStructureID:     d.PrizeStructureID.String(),
			Status:               d.Status,
			TotalEligibleMSISDNs: d.TotalEligibleMSISDNs,
			TotalEntries:         d.TotalEntries,
			ExecutedByAdminID:    d.ExecutedByAdminID.String(),
			Winners:              winners,
			CreatedAt:            d.CreatedAt.Format(time.RFC3339),
		}
		draws = append(draws, draw)
	}

	// Prepare response
	resp := response.PaginatedResponse{
		Success: true,
		Data:    draws,
		Pagination: response.Pagination{
			Page:      page,
			PageSize:  pageSize,
			TotalRows: output.Total,
			TotalPages: (output.Total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetEligibilityStats handles the request to get eligibility statistics for a draw date
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	// Parse draw date
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Missing date parameter",
			Details: "Date parameter is required",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid date format",
			Details: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	// Get eligibility stats
	input := draw.GetEligibilityStatsInput{
		Date: date,
	}

	output, err := h.getEligibilityStats.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get eligibility statistics",
			Details: err.Error(),
		})
		return
	}

	// Prepare response
	resp := response.EligibilityStatsResponse{
		Date:                 date.Format("2006-01-02"),
		TotalEligibleMSISDNs: output.TotalEligibleMSISDNs,
		TotalEntries:         output.TotalEntries,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// InvokeRunnerUp handles the request to invoke a runner-up for a prize
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	var req request.InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid winner ID",
			Details: err.Error(),
		})
		return
	}

	// Parse admin ID from JWT token
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User ID not found in token",
		})
		return
	}
	
	adminID, ok := adminIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Internal server error",
			Details: "Failed to parse user ID",
		})
		return
	}
	
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Internal server error",
			Details: "Invalid user ID format",
		})
		return
	}

	// Invoke runner-up
	input := draw.InvokeRunnerUpInput{
		WinnerID:    winnerID,
		Reason:      req.Reason,
		AdminID:     adminUUID,
	}

	output, err := h.invokeRunnerUp.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.DrawError:
			drawErr := err.(*entity.DrawError)
			switch drawErr.Code() {
			case entity.ErrWinnerNotFound:
				statusCode = http.StatusNotFound
				errorMessage = "Winner not found"
			case entity.ErrNoRunnerUpsAvailable:
				statusCode = http.StatusBadRequest
				errorMessage = "No runner-ups available for this prize"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to invoke runner-up"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to invoke runner-up"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Prepare response
	resp := response.RunnerUpResponse{
		OriginalWinner: response.WinnerResponse{
			ID:            output.OriginalWinner.ID.String(),
			MSISDN:        maskMSISDN(output.OriginalWinner.MSISDN),
			PrizeTierID:   output.OriginalWinner.PrizeTierID.String(),
			PrizeTierName: output.OriginalWinner.PrizeTierName,
			PrizeValue:    output.OriginalWinner.PrizeValue,
			Status:        output.OriginalWinner.Status,
			IsRunnerUp:    output.OriginalWinner.IsRunnerUp,
			RunnerUpRank:  output.OriginalWinner.RunnerUpRank,
		},
		NewWinner: response.WinnerResponse{
			ID:            output.NewWinner.ID.String(),
			MSISDN:        maskMSISDN(output.NewWinner.MSISDN),
			PrizeTierID:   output.NewWinner.PrizeTierID.String(),
			PrizeTierName: output.NewWinner.PrizeTierName,
			PrizeValue:    output.NewWinner.PrizeValue,
			Status:        output.NewWinner.Status,
			IsRunnerUp:    output.NewWinner.IsRunnerUp,
			RunnerUpRank:  output.NewWinner.RunnerUpRank,
		},
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// Helper function to mask MSISDN
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 6 {
		return msisdn
	}
	
	first3 := msisdn[:3]
	last3 := msisdn[len(msisdn)-3:]
	masked := first3 + "****" + last3
	
	return masked
}
