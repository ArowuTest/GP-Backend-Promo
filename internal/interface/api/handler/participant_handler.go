package interface

import (
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant/entity"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ParticipantHandler handles HTTP requests related to participants
type ParticipantHandler struct {
	uploadParticipants *participant.UploadParticipantsUseCase
	listParticipants   *participant.ListParticipantsUseCase
	getParticipantStats *participant.GetParticipantStatsUseCase
	deleteUpload       *participant.DeleteUploadUseCase
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	uploadParticipants *participant.UploadParticipantsUseCase,
	listParticipants *participant.ListParticipantsUseCase,
	getParticipantStats *participant.GetParticipantStatsUseCase,
	deleteUpload *participant.DeleteUploadUseCase,
) *ParticipantHandler {
	return &ParticipantHandler{
		uploadParticipants: uploadParticipants,
		listParticipants:   listParticipants,
		getParticipantStats: getParticipantStats,
		deleteUpload:       deleteUpload,
	}
}

// UploadParticipants handles the request to upload participant data
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to parse form",
			Details: err.Error(),
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get file",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// Check file type
	if header.Header.Get("Content-Type") != "text/csv" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid file type",
			Details: "Only CSV files are supported",
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

	// Parse CSV file
	// In a real implementation, we would read the file and parse the CSV data
	// For this example, we'll create some dummy data
	participants := []participant.ParticipantData{
		{
			MSISDN:         "2347012345678",
			RechargeAmount: 500.0,
			RechargeDate:   time.Now(),
		},
		{
			MSISDN:         "2347087654321",
			RechargeAmount: 1000.0,
			RechargeDate:   time.Now(),
		},
	}

	// Upload participants
	input := participant.UploadParticipantsInput{
		FileName:     header.Filename,
		UploadedBy:   adminUUID,
		Participants: participants,
	}

	output, err := h.uploadParticipants.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.ParticipantError:
			participantErr := err.(*entity.ParticipantError)
			switch participantErr.Code() {
			case entity.ErrInvalidMSISDN:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid MSISDN format"
			case entity.ErrDuplicateParticipant:
				statusCode = http.StatusBadRequest
				errorMessage = "Duplicate participant entries"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to upload participants"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to upload participants"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Prepare response
	resp := response.UploadResponse{
		AuditID:           output.AuditID.String(),
		Status:            output.Status,
		TotalRowsProcessed: output.TotalRowsProcessed,
		SuccessfulRows:    output.SuccessfulRows,
		ErrorCount:        output.ErrorCount,
		ErrorDetails:      output.ErrorDetails,
		DuplicatesSkipped: output.DuplicatesSkipped,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// ListParticipants handles the request to list participants
func (h *ParticipantHandler) ListParticipants(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Parse date filter
	dateStr := c.Query("date")
	var date *time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
		date = &parsedDate
	}

	// List participants
	input := participant.ListParticipantsInput{
		Page:     page,
		PageSize: pageSize,
		Date:     date,
	}

	output, err := h.listParticipants.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list participants",
			Details: err.Error(),
		})
		return
	}

	// Convert participants to response format
	participants := make([]response.ParticipantResponse, 0, len(output.Participants))
	for _, p := range output.Participants {
		participant := response.ParticipantResponse{
			ID:             p.ID.String(),
			MSISDN:         maskMSISDN(p.MSISDN),
			Points:         p.Points,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   p.RechargeDate.Format("2006-01-02"),
			CreatedAt:      p.CreatedAt.Format(time.RFC3339),
		}
		participants = append(participants, participant)
	}

	// Prepare response
	resp := response.PaginatedResponse{
		Success: true,
		Data:    participants,
		Pagination: response.Pagination{
			Page:      page,
			PageSize:  pageSize,
			TotalRows: output.Total,
			TotalPages: (output.Total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetParticipantStats handles the request to get participant statistics
func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	// Parse date
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

	// Get participant stats
	input := participant.GetParticipantStatsInput{
		Date: date,
	}

	output, err := h.getParticipantStats.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get participant statistics",
			Details: err.Error(),
		})
		return
	}

	// Prepare response
	resp := response.ParticipantStatsResponse{
		Date:              date.Format("2006-01-02"),
		TotalParticipants: output.TotalParticipants,
		TotalPoints:       output.TotalPoints,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// DeleteUpload handles the request to delete an upload
func (h *ParticipantHandler) DeleteUpload(c *gin.Context) {
	// Parse upload ID
	uploadIDStr := c.Param("id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid upload ID",
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

	// Delete upload
	input := participant.DeleteUploadInput{
		UploadID: uploadID,
		AdminID:  adminUUID,
	}

	err = h.deleteUpload.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.ParticipantError:
			participantErr := err.(*entity.ParticipantError)
			if participantErr.Code() == entity.ErrUploadNotFound {
				statusCode = http.StatusNotFound
				errorMessage = "Upload not found"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to delete upload"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to delete upload"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    "Upload deleted successfully",
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
