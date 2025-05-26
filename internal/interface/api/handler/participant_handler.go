package handler

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	participantApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	uploadParticipantsService *participantApp.UploadParticipantsService
	getParticipantService     *participantApp.GetParticipantService
	listParticipantsService   *participantApp.ListParticipantsService
	getParticipantStatsService *participantApp.GetParticipantStatsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	uploadParticipantsService *participantApp.UploadParticipantsService,
	getParticipantService *participantApp.GetParticipantService,
	listParticipantsService *participantApp.ListParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
) *ParticipantHandler {
	return &ParticipantHandler{
		uploadParticipantsService: uploadParticipantsService,
		getParticipantService:     getParticipantService,
		listParticipantsService:   listParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
	}
}

// UploadParticipants handles POST /api/admin/participants/upload
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	var req request.UploadParticipantsRequest
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
	participants := make([]participantApp.ParticipantInput, 0, len(req.Participants))
	for _, p := range req.Participants {
		participants = append(participants, participantApp.ParticipantInput{
			MSISDN:         p.MSISDN,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   p.RechargeDate,
		})
	}
	
	input := participantApp.UploadParticipantsInput{
		Participants: participants,
		UploadedBy:   userID.(uuid.UUID),
	}
	
	// Upload participants
	output, err := h.uploadParticipantsService.UploadParticipants(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to upload participants: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.UploadParticipantsResponse{
			TotalUploaded: output.TotalUploaded,
			UploadID:      output.UploadID.String(),
			UploadedAt:    output.UploadedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// GetParticipant handles GET /api/admin/participants/:id
func (h *ParticipantHandler) GetParticipant(c *gin.Context) {
	// Parse participant ID
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid participant ID format",
		})
		return
	}
	
	// Prepare input
	input := participantApp.GetParticipantInput{
		ID: participantID,
	}
	
	// Get participant
	output, err := h.getParticipantService.GetParticipant(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get participant: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantResponse{
			ID:             output.ID.String(),
			MSISDN:         output.MSISDN,
			Points:         output.Points,
			RechargeAmount: output.RechargeAmount,
			RechargeDate:   output.RechargeDate,
			CreatedAt:      output.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      output.UpdatedAt.Format("2006-01-02 15:04:05"),
			UploadID:       output.UploadID.String(),
			UploadedAt:     output.UploadedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// ListParticipants handles GET /api/admin/participants
func (h *ParticipantHandler) ListParticipants(c *gin.Context) {
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
	input := participantApp.ListParticipantsInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	// List participants
	output, err := h.listParticipantsService.ListParticipants(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list participants: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	participants := make([]response.ParticipantResponse, 0, len(output.Participants))
	for _, p := range output.Participants {
		participants = append(participants, response.ParticipantResponse{
			ID:             p.ID.String(),
			MSISDN:         p.MSISDN,
			Points:         p.Points,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   p.RechargeDate,
			CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
			UploadID:       p.UploadID.String(),
			UploadedAt:     p.UploadedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    participants,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// GetParticipantStats handles GET /api/admin/participants/stats
func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	// Parse date parameter
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	
	// Prepare input
	input := participantApp.GetParticipantStatsInput{
		Date: date,
	}
	
	// Get participant stats
	output, err := h.getParticipantStatsService.GetParticipantStats(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get participant stats: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			Date:              output.Date,
			TotalParticipants: output.TotalParticipants,
			TotalPoints:       output.TotalPoints,
			AveragePoints:     float64(output.TotalPoints) / float64(output.TotalParticipants),
		},
	})
}
