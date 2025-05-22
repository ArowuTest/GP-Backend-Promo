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
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	listParticipantsService    *participantApp.ListParticipantsService
	getParticipantStatsService *participantApp.GetParticipantStatsService
	listUploadAuditsService    *participantApp.ListUploadAuditsService
	uploadParticipantsService  *participantApp.UploadParticipantsService
	deleteUploadService        *participantApp.DeleteUploadService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	listParticipantsService *participantApp.ListParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
	listUploadAuditsService *participantApp.ListUploadAuditsService,
	uploadParticipantsService *participantApp.UploadParticipantsService,
	deleteUploadService *participantApp.DeleteUploadService,
) *ParticipantHandler {
	return &ParticipantHandler{
		listParticipantsService:    listParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
		listUploadAuditsService:    listUploadAuditsService,
		uploadParticipantsService:  uploadParticipantsService,
		deleteUploadService:        deleteUploadService,
	}
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
			RechargeDate:   util.FormatTimeOrEmpty(p.RechargeDate, "2006-01-02"),
			CreatedAt:      util.FormatTimeOrEmpty(p.CreatedAt, time.RFC3339),
			// Note: UpdatedAt field is not in ParticipantResponse DTO
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
	// Parse date parameters
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	
	// Prepare input with required fields from the application layer
	input := participantApp.GetParticipantStatsInput{
		StartDate: startDate,
		EndDate:   endDate,
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
	
	// Prepare response - only use fields that exist in ParticipantStatsResponse DTO
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			Date:              output.StartDate, // Use StartDate as Date
			TotalParticipants: output.TotalParticipants,
			TotalPoints:       output.TotalPoints,
			// Note: AveragePoints, StartDate, EndDate fields are not in ParticipantStatsResponse DTO
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
	
	// Prepare response - output.Audits is the correct field name in ListUploadAuditsOutput
	audits := make([]response.UploadAuditResponse, 0, len(output.Audits))
	for _, a := range output.Audits {
		// Parse error details string to slice
		errorDetails := []string{}
		if a.ErrorDetailsStr != "" {
			errorDetails = append(errorDetails, a.ErrorDetailsStr)
		}
		
		audits = append(audits, response.UploadAuditResponse{
			ID:             a.ID.String(),
			UploadedBy:     a.UploadedBy.String(),
			UploadDate:     util.FormatTimeOrEmpty(a.UploadDate, time.RFC3339), // Using UploadDate from domain entity
			FileName:       a.FileName,
			Status:         a.Status,
			TotalRows:      a.TotalRows,
			SuccessfulRows: a.SuccessfulRows,
			ErrorCount:     a.ErrorCount, // Using ErrorCount instead of ErrorRows
			ErrorDetails:   errorDetails,
			// Note: CreatedAt and UpdatedAt fields are not in UploadAuditResponse DTO
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    audits,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
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
	
	// Prepare input - using only fields that exist in the application layer ParticipantInput
	participants := make([]participantApp.ParticipantInput, 0, len(req.Participants))
	for _, p := range req.Participants {
		participants = append(participants, participantApp.ParticipantInput{
			MSISDN:         p.MSISDN,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   p.RechargeDate, // This is a string in both request and application layer
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
	
	// Prepare response - using only fields that exist in the response DTO
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UploadResponse{
			AuditID:           output.UploadID.String(), // Map UploadID to AuditID
			Status:            "Completed",
			TotalRowsProcessed: output.TotalUploaded,
			SuccessfulRows:    output.TotalUploaded,
			ErrorCount:        0, // Default value
			ErrorDetails:      []string{},
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
	
	// Create DeleteUploadInput struct instead of passing UUID directly
	input := participantApp.DeleteUploadInput{
		UploadID: uploadID,
	}
	
	// Delete upload - capture both return values
	result, err := h.deleteUploadService.DeleteUpload(c.Request.Context(), input)
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
		Data: response.DeleteConfirmationResponse{
			ID:      uploadID.String(),
			Deleted: result,
		},
	})
}
