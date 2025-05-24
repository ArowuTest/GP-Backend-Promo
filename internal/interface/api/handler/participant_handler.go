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

	// Parse search query
	search := c.DefaultQuery("search", "")

	// Prepare input
	input := participantApp.ListParticipantsInput{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
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

	// Prepare response with explicit type conversions at DTO boundary
	participants := make([]response.ParticipantResponse, 0, len(output.Participants))
	for _, p := range output.Participants {
		participants = append(participants, response.ParticipantResponse{
			ID:        p.ID,
			MSISDN:    p.MSISDN,
			Points:    p.Points,
			CreatedAt: util.FormatTimeOrEmpty(p.CreatedAt, time.RFC3339),
			UpdatedAt: util.FormatTimeOrEmpty(p.UpdatedAt, time.RFC3339),
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
	// Get participant stats
	output, err := h.getParticipantStatsService.GetParticipantStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get participant stats: " + err.Error(),
		})
		return
	}

	// Prepare response with explicit type conversions at DTO boundary
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			TotalParticipants: output.TotalParticipants,
			TotalPoints:       output.TotalPoints,
			AveragePoints:     output.AveragePoints,
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
	
	// Prepare response with fields that exist in UploadAuditResponse DTO
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
			UploadDate:     util.FormatTimeOrEmpty(a.CreatedAt, time.RFC3339),
			FileName:       a.FileName,
			Status:         a.Status,
			TotalRows:      a.TotalRows,
			SuccessfulRows: a.SuccessfulRows,
			ErrorCount:     len(errorDetails),
			ErrorDetails:   errorDetails,
			// Additional fields for frontend compatibility
			UploadedAt:    util.FormatTimeOrEmpty(a.CreatedAt, time.RFC3339),
			TotalUploaded: a.TotalRows,
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
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Type assertion with safety check
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try to parse as string if not UUID
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Success: false,
					Error:   "Invalid user ID format in token",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID type in token",
			})
			return
		}
	}
	
	// Prepare input with fields that exist in ParticipantInput
	participants := make([]participantApp.ParticipantInput, 0, len(req.Participants))
	for _, p := range req.Participants {
		participants = append(participants, participantApp.ParticipantInput{
			MSISDN: p.MSISDN,
			Points: p.Points,
		})
	}
	
	input := participantApp.UploadParticipantsInput{
		Participants: participants,
		UploadedBy:   userID,
		FileName:     req.FileName,
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
	
	// Prepare response with fields that exist in UploadResponse DTO
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UploadResponse{
			AuditID:           output.UploadID.String(),
			Status:            "Completed",
			TotalRowsProcessed: output.TotalUploaded,
			SuccessfulRows:    output.SuccessfulUploaded,
			ErrorCount:        output.ErrorsEncountered,
			ErrorDetails:      []string{},
			DuplicatesSkipped: output.DuplicatesSkipped,
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
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Type assertion with safety check
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try to parse as string if not UUID
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Success: false,
					Error:   "Invalid user ID format in token",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID type in token",
			})
			return
		}
	}
	
	// Create DeleteUploadInput struct
	input := participantApp.DeleteUploadInput{
		UploadID:  uploadID,
		DeletedBy: userID,
	}
	
	// Delete upload and capture both return values
	output, err := h.deleteUploadService.DeleteUpload(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete upload: " + err.Error(),
		})
		return
	}
	
	// Prepare response with DeleteConfirmationResponse DTO
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DeleteConfirmationResponse{
			ID:      uploadID.String(),
			Deleted: output.Deleted,
		},
	})
}
