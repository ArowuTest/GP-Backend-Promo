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
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	listParticipantsService *participantApp.ListParticipantsService
	listUploadAuditsService *participantApp.ListUploadAuditsService
	deleteUploadService     *participantApp.DeleteUploadService
	uploadParticipantsService *participantApp.UploadParticipantsService
	getParticipantStatsService *participantApp.GetParticipantStatsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	listParticipantsService *participantApp.ListParticipantsService,
	listUploadAuditsService *participantApp.ListUploadAuditsService,
	deleteUploadService *participantApp.DeleteUploadService,
	uploadParticipantsService *participantApp.UploadParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
) *ParticipantHandler {
	return &ParticipantHandler{
		listParticipantsService: listParticipantsService,
		listUploadAuditsService: listUploadAuditsService,
		deleteUploadService: deleteUploadService,
		uploadParticipantsService: uploadParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
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
			Details: "An error occurred while retrieving participants. Please try again later.",
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
			CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
			UploadID:       p.UploadID.String(),
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

// UploadParticipants handles POST /api/admin/participants/upload
func (h *ParticipantHandler) UploadParticipants(c *gin.Context) {
	// Check if this is a JSON request or a multipart form
	contentType := c.GetHeader("Content-Type")
	if contentType == "application/json" {
		// Handle JSON request
		var req request.UploadParticipantsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
				Details: "Please check the format of your upload data. Ensure all required fields are present and correctly formatted.",
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
			FileName:     req.FileName,
		}
		
		// Upload participants
		output, err := h.uploadParticipantsService.UploadParticipants(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Failed to upload participants: " + err.Error(),
				Details: "An error occurred while processing your upload. Please try again later.",
			})
			return
		}
		
		// Prepare response
		c.JSON(http.StatusOK, response.SuccessResponse{
			Success: true,
			Data: response.UploadResponse{
				AuditID:           output.AuditID.String(),
				Status:            output.Status,
				TotalRowsProcessed: output.TotalRowsProcessed,
				SuccessfulRows:    output.SuccessfulRows,
				ErrorCount:        output.ErrorCount,
				ErrorDetails:      output.ErrorDetails,
				DuplicatesSkipped: output.DuplicatesSkipped,
			},
		})
	} else {
		// Handle multipart form (CSV upload)
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
				Details: "An error occurred while processing your file. Please try again later.",
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
				Details: "The CSV file appears to be empty or corrupted.",
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
					Details: "Please check your CSV file for formatting errors.",
				})
				return
			}
			
			// Parse recharge amount
			rechargeAmount, err := strconv.ParseFloat(record[rechargeAmountIdx], 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("Invalid recharge amount at line %d: %s", lineNum, err.Error()),
					Details: "Recharge amount must be a valid number.",
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
			FileName:     file.Filename,
		}
		
		// Upload participants
		output, err := h.uploadParticipantsService.UploadParticipants(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Failed to upload participants: " + err.Error(),
				Details: "An error occurred while processing your upload. Please try again later.",
			})
			return
		}
		
		// Prepare response
		c.JSON(http.StatusOK, response.SuccessResponse{
			Success: true,
			Data: response.UploadParticipantsResponse{
				TotalUploaded: output.TotalRowsProcessed,
				UploadID:      output.AuditID.String(),
				UploadedAt:    time.Now().Format("2006-01-02 15:04:05"),
				FileName:      file.Filename,
				SuccessfullyImported: output.SuccessfulRows,
				DuplicatesSkipped: output.DuplicatesSkipped,
				ErrorsEncountered: output.ErrorCount,
				Status: output.Status,
				Notes: output.ErrorDetails,
				OperationType: "Upload",
			},
		})
	}
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
			Details: "An error occurred while retrieving participant statistics. Please try again later.",
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
			Details: "An error occurred while retrieving upload history. Please try again later.",
		})
		return
	}
	
	// Prepare response
	audits := make([]response.UploadAuditResponse, 0, len(output.Audits))
	for _, a := range output.Audits {
		audits = append(audits, response.UploadAuditResponse{
			ID:             a.ID.String(),
			UploadedBy:     a.UploadedBy.String(),
			UploadDate:     a.UploadDate.Format("2006-01-02"),
			FileName:       a.FileName,
			Status:         a.Status,
			TotalRows:      a.TotalRows,
			SuccessfulRows: a.SuccessfulRows,
			ErrorCount:     a.ErrorCount,
			ErrorDetails:   a.ErrorDetails,
			UploadedAt:     a.UploadDate.Format("2006-01-02 15:04:05"),
			TotalUploaded:  a.TotalRows,
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

// DeleteUpload handles DELETE /api/admin/participants/uploads/:id
func (h *ParticipantHandler) DeleteUpload(c *gin.Context) {
	// Parse upload ID
	uploadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid upload ID format",
			Details: "The provided ID is not in the correct UUID format",
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
			Details: "An error occurred while deleting the upload. Please try again later.",
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.DeleteConfirmationResponse{
			ID:      output.UploadID.String(),
			Deleted: output.Deleted,
		},
	})
}
