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
	uploadParticipantsService   *participantApp.UploadParticipantsService
	listParticipantsService     *participantApp.ListParticipantsService
	getParticipantStatsService  *participantApp.GetParticipantStatsService
	listUploadAuditsService     *participantApp.ListUploadAuditsService
	deleteUploadService         *participantApp.DeleteUploadService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	uploadParticipantsService *participantApp.UploadParticipantsService,
	listParticipantsService *participantApp.ListParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
	listUploadAuditsService *participantApp.ListUploadAuditsService,
	deleteUploadService *participantApp.DeleteUploadService,
) *ParticipantHandler {
	return &ParticipantHandler{
		uploadParticipantsService:   uploadParticipantsService,
		listParticipantsService:     listParticipantsService,
		getParticipantStatsService:  getParticipantStatsService,
		listUploadAuditsService:     listUploadAuditsService,
		deleteUploadService:         deleteUploadService,
	}
}

// UploadParticipants handles POST /api/admin/participants/upload
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "No file uploaded: " + err.Error(),
			Details: "Please upload a CSV file with participant data",
		})
		return
	}
	defer file.Close()
	
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
	input := participantApp.UploadParticipantsInput{
		File:       file,
		FileName:   header.Filename,
		UploadedBy: userID.(uuid.UUID),
	}
	
	// Upload participants
	output, err := h.uploadParticipantsService.UploadParticipants(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to upload participants: " + err.Error(),
			Details: err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UploadParticipantsResponse{
			TotalUploaded:        output.TotalUploaded,
			UploadID:             output.UploadID.String(),
			UploadedAt:           output.UploadedAt.Format(time.RFC3339),
			FileName:             output.FileName,
			SuccessfullyImported: output.SuccessfullyImported,
			DuplicatesSkipped:    output.DuplicatesSkipped,
			ErrorsEncountered:    output.ErrorsEncountered,
			Status:               output.Status,
			Notes:                output.Notes,
			OperationType:        "PARTICIPANT_UPLOAD",
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
			RechargeDate:   p.RechargeDate.Format("2006-01-02"),
			CreatedAt:      p.CreatedAt.Format(time.RFC3339),
			UploadID:       p.UploadID.String(),
			UploadedAt:     p.UploadedAt.Format(time.RFC3339),
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
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			Date:              output.Date,
			TotalParticipants: output.TotalParticipants,
			TotalPoints:       output.TotalPoints,
		},
	})
}

// ListUploadAudits handles GET /api/admin/participants/uploads
func (h *ParticipantHandler) ListUploadAudits(c *gin.Context) {
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
	input := participantApp.ListUploadAuditsInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	// List upload audits
	output, err := h.listUploadAuditsService.ListUploadAudits(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list upload audits: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	uploadAudits := make([]response.UploadAuditResponse, 0, len(output.UploadAudits))
	for _, ua := range output.UploadAudits {
		uploadAudits = append(uploadAudits, response.UploadAuditResponse{
			ID:             ua.ID.String(),
			UploadedBy:     ua.UploadedBy.String(),
			UploadDate:     ua.UploadDate.Format(time.RFC3339),
			FileName:       ua.FileName,
			Status:         ua.Status,
			TotalRows:      ua.TotalRows,
			SuccessfulRows: ua.SuccessfulRows,
			ErrorCount:     ua.ErrorCount,
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    uploadAudits,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// DeleteUpload handles DELETE /api/admin/participants/uploads/:id
func (h *ParticipantHandler) DeleteUpload(c *gin.Context) {
	// Parse upload ID
	uploadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid upload ID format",
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
	input := participantApp.DeleteUploadInput{
		UploadID:  uploadID,
		DeletedBy: userID.(uuid.UUID),
	}
	
	// Delete upload
	output, err := h.deleteUploadService.DeleteUpload(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete upload: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":      output.UploadID.String(),
			"deleted": true,
		},
	})
}
