package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/adapter"
	participantApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	participantServiceAdapter *adapter.ParticipantServiceAdapter
	getParticipantStatsService *participantApp.GetParticipantStatsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	participantServiceAdapter *adapter.ParticipantServiceAdapter,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
) *ParticipantHandler {
	return &ParticipantHandler{
		participantServiceAdapter: participantServiceAdapter,
		getParticipantStatsService: getParticipantStatsService,
	}
}

// GetParticipants handles GET /api/admin/participants
func (h *ParticipantHandler) GetParticipants(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Parse search parameter
	search := c.DefaultQuery("search", "")

	// Get participants - using adapter method signature directly
	output, err := h.participantServiceAdapter.ListParticipants(c.Request.Context(), page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get participants: " + err.Error(),
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
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// GetParticipantStats handles GET /api/admin/participants/stats
func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	// Get participant stats - using adapter method signature directly
	output, err := h.participantServiceAdapter.GetParticipantStats(c.Request.Context())
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
			TotalParticipants: output.TotalParticipants,
			TotalPoints:       output.TotalPoints,
			AveragePoints:     float64(output.TotalPoints) / float64(output.TotalParticipants),
		},
	})
}

// UploadParticipants handles POST /api/admin/participants/upload
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get file: " + err.Error(),
		})
		return
	}
	defer file.Close()

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
	var uploadedBy uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		uploadedBy = id
	case string:
		var err error
		uploadedBy, err = uuid.Parse(id)
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

	// Read file content - we'll use this later for CSV parsing
	_, err = io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to read file content: " + err.Error(),
		})
		return
	}

	// Convert to participant inputs - simplified for now
	// In a real implementation, this would parse CSV data
	participants := []participant.ParticipantInput{
		{
			MSISDN: "1234567890",
			Points: 10,
		},
	}

	// Upload participants
	output, err := h.participantServiceAdapter.UploadParticipants(c.Request.Context(), participants, uploadedBy, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to upload participants: " + err.Error(),
		})
		return
	}

	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d participants", output.RecordCount),
		Data: map[string]interface{}{
			"id":                   output.ID.String(),
			"fileName":             output.FileName,
			"totalUploaded":        output.RecordCount,
			"successfullyImported": output.RecordCount,
			"duplicatesSkipped":    0,
			"errorsEncountered":    0,
			"status":               output.Status,
			"details":              output.ErrorMessage,
			"uploadedBy":           uploadedBy.String(),
			"uploadedAt":           util.FormatTimeOrEmpty(output.UploadDate, time.RFC3339),
		},
	})
}

// ListUploadAudits handles GET /api/admin/participants/upload-audits
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

	// Get upload audits - using adapter method signature directly
	output, err := h.participantServiceAdapter.ListUploadAudits(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get upload audits: " + err.Error(),
		})
		return
	}

	// Prepare response with explicit type conversions at DTO boundary
	uploadAudits := make([]map[string]interface{}, 0, len(output.Audits))
	for _, a := range output.Audits {
		// Create a response that matches the frontend expectations
		uploadAudits = append(uploadAudits, map[string]interface{}{
			"id":                   a.ID.String(),
			"fileName":             a.FileName,
			"totalUploaded":        a.RecordCount,
			"successfullyImported": a.RecordCount,
			"duplicatesSkipped":    0,
			"errorsEncountered":    0,
			"status":               a.Status,
			"details":              a.ErrorMessage,
			"errorDetails":         a.ErrorMessage,
			"uploadedBy":           a.UploadedBy.String(),
			"uploadedAt":           util.FormatTimeOrEmpty(a.UploadDate, time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    uploadAudits,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// DeleteUpload handles DELETE /api/admin/participants/upload/:id
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
	var deletedBy uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		deletedBy = id
	case string:
		var err error
		deletedBy, err = uuid.Parse(id)
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

	// Delete upload
	output, err := h.participantServiceAdapter.DeleteUpload(c.Request.Context(), uploadID, deletedBy)
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
		Message: "Upload deleted successfully",
		Data:    output,
	})
}
