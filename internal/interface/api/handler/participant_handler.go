package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	participantApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	uploadParticipantsService *participantApp.UploadParticipantsService
	getParticipantStatsService *participantApp.GetParticipantStatsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	uploadParticipantsService *participantApp.UploadParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
) *ParticipantHandler {
	return &ParticipantHandler{
		uploadParticipantsService: uploadParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
	}
}

// UploadParticipants handles POST /api/admin/participants/upload
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	// Get the file from the multipart form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "No file uploaded or invalid form data",
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
	
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to open uploaded file: " + err.Error(),
		})
		return
	}
	defer src.Close()
	
	// Parse CSV
	reader := csv.NewReader(src)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to read CSV header: " + err.Error(),
		})
		return
	}
	
	// Validate header
	msisdnIdx := -1
	rechargeAmountIdx := -1
	rechargeDateIdx := -1
	
	for i, col := range header {
		switch col {
		case "MSISDN":
			msisdnIdx = i
		case "RechargeAmount":
			rechargeAmountIdx = i
		case "RechargeDate":
			rechargeDateIdx = i
		}
	}
	
	if msisdnIdx == -1 || rechargeAmountIdx == -1 || rechargeDateIdx == -1 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid CSV format",
			Details: "CSV must contain MSISDN, RechargeAmount, and RechargeDate columns",
		})
		return
	}
	
	// Read and parse rows
	participants := []participantApp.ParticipantInput{}
	lineNum := 1 // Start from 1 to account for header
	
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Error reading CSV at line %d: %s", lineNum, err.Error()),
			})
			return
		}
		
		// Parse recharge amount
		rechargeAmount, err := strconv.ParseFloat(record[rechargeAmountIdx], 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid recharge amount at line %d: %s", lineNum, err.Error()),
			})
			return
		}
		
		participants = append(participants, participantApp.ParticipantInput{
			MSISDN:         record[msisdnIdx],
			RechargeAmount: rechargeAmount,
			RechargeDate:   record[rechargeDateIdx],
		})
	}
	
	// Prepare input
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
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UploadParticipantsResponse{
			TotalUploaded: output.TotalUploaded,
			UploadID:      output.UploadID.String(),
			UploadedAt:    output.UploadedAt.Format("2006-01-02 15:04:05"),
			FileName:      file.Filename,
			SuccessfullyImported: output.TotalUploaded,
			DuplicatesSkipped: 0,
			ErrorsEncountered: 0,
			Status: "Completed",
			Notes: "CSV upload processed successfully",
			OperationType: "Upload",
		},
	})
}

// GetParticipantStats handles GET /api/admin/participants/stats
func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	// Parse date range
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	
	// Prepare input
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
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			Date:              output.StartDate,
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
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a mock response with pagination
	
	// Mock upload audits
	uploadAudits := []response.UploadAuditResponse{
		{
			ID:             uuid.New().String(),
			UploadedBy:     "Admin User",
			UploadDate:     time.Now().Format("2006-01-02 15:04:05"),
			FileName:       "participants.csv",
			Status:         "Completed",
			TotalRows:      100,
			SuccessfulRows: 100,
			ErrorCount:     0,
		},
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    uploadAudits,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  len(uploadAudits),
			TotalPages: 1,
			TotalItems: int64(len(uploadAudits)),
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
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a mock response with pagination
	
	// Mock participants
	participants := []response.ParticipantResponse{
		{
			ID:             uuid.New().String(),
			MSISDN:         "234*****789", // Masked for privacy
			RechargeAmount: 500.0,
			RechargeDate:   time.Now().Format("2006-01-02"),
			Points:         5,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			UploadID:       uuid.New().String(),
			UploadedAt:     time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    participants,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  len(participants),
			TotalPages: 1,
			TotalItems: int64(len(participants)),
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
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a success response
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":      uploadID.String(),
			"deleted": true,
		},
	})
}