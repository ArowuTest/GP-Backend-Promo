package handler

import (
	"encoding/base64"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
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
	listParticipantsService   *participantApp.ListParticipantsService
	getParticipantStatsService *participantApp.GetParticipantStatsService
	listUploadAuditsService   *participantApp.ListUploadAuditsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(
	uploadParticipantsService *participantApp.UploadParticipantsService,
	listParticipantsService *participantApp.ListParticipantsService,
	getParticipantStatsService *participantApp.GetParticipantStatsService,
	listUploadAuditsService *participantApp.ListUploadAuditsService,
) *ParticipantHandler {
	return &ParticipantHandler{
		uploadParticipantsService: uploadParticipantsService,
		listParticipantsService:   listParticipantsService,
		getParticipantStatsService: getParticipantStatsService,
		listUploadAuditsService:   listUploadAuditsService,
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
	
	// Parse CSV data
	csvData, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid CSV data: " + err.Error(),
		})
		return
	}
	
	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to parse CSV: " + err.Error(),
		})
		return
	}
	
	// Skip header row
	if len(records) > 0 {
		records = records[1:]
	}
	
	// Prepare input
	participants := make([]participantApp.ParticipantInput, 0, len(records))
	for _, record := range records {
		if len(record) < 3 {
			continue
		}
		
		rechargeAmount, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}
		
		participants = append(participants, participantApp.ParticipantInput{
			MSISDN:         record[0],
			RechargeAmount: rechargeAmount,
			RechargeDate:   record[2],
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
		})
		return
	}
	
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.UploadParticipantsResponse{
			TotalUploaded: output.TotalUploaded,
			UploadID:      output.UploadID.String(),
			UploadedAt:    output.UploadDate.Format("2006-01-02 15:04:05"),
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
			CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
			UploadID:       p.UploadID.String(),
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
	// Parse date parameters
	startDate := c.DefaultQuery("startDate", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("endDate", time.Now().Format("2006-01-02"))
	
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
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.ParticipantStatsResponse{
			Date:              output.StartDate,
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
	
	// Prepare response
	audits := make([]response.DataUploadAuditResponse, 0, len(output.Audits))
	for _, a := range output.Audits {
		// Get user name from ID (in a real implementation, this would be a service call)
		userName := "Admin User" // Placeholder
		
		audits = append(audits, response.DataUploadAuditResponse{
			ID:                  a.ID.String(),
			UploadedBy:          userName,
			UploadedByUserId:    a.UploadedBy.String(),
			UploadedAt:          a.CreatedAt.Format("2006-01-02 15:04:05"),
			TotalUploaded:       a.TotalUploaded,
			Status:              a.Status,
			SuccessfullyImported: a.SuccessfulRows,
			ErrorsEncountered:   a.ErrorCount,
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    audits,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}
