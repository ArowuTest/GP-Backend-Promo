package handler

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	auditApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	getAuditLogsService *auditApp.GetAuditLogsService
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(
	getAuditLogsService *auditApp.GetAuditLogsService,
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService,
) *AuditHandler {
	return &AuditHandler{
		getAuditLogsService: getAuditLogsService,
		getDataUploadAuditsService: getDataUploadAuditsService,
	}
}

// GetAuditLogs handles GET /api/admin/audit-logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	var req request.GetAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check your query parameters and try again.",
		})
		return
	}
	
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Parse filter parameters
	action := c.Query("action")
	entityType := c.Query("entityType")
	
	var entityID uuid.UUID
	if entityIDStr := c.Query("entityId"); entityIDStr != "" {
		entityID, err = uuid.Parse(entityIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid entity ID format",
				Details: "The provided entity ID is not in the correct UUID format",
			})
			return
		}
	}
	
	var userID uuid.UUID
	if userIDStr := c.Query("userId"); userIDStr != "" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format",
				Details: "The provided user ID is not in the correct UUID format",
			})
			return
		}
	}
	
	// Parse date range
	startDate, _ := parseDate(c.Query("startDate"))
	endDate, _ := parseDate(c.Query("endDate"))
	
	// Prepare input
	input := auditApp.GetAuditLogsInput{
		Page:     page,
		PageSize: pageSize,
		Filters: auditApp.AuditLogFilters{
			Action:     action,
			EntityType: entityType,
			EntityID:   entityID,
			UserID:     userID,
			StartDate:  startDate,
			EndDate:    endDate,
		},
	}
	
	// Get audit logs
	output, err := h.getAuditLogsService.GetAuditLogs(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get audit logs: " + err.Error(),
			Details: "An error occurred while retrieving audit logs. Please try again later.",
		})
		return
	}
	
	// Prepare response
	auditLogs := make([]response.AuditLogResponse, 0, len(output.AuditLogs))
	for _, log := range output.AuditLogs {
		auditLogs = append(auditLogs, response.AuditLogResponse{
			ID:         log.ID.String(),
			UserID:     log.UserID.String(),
			Username:   log.Username,
			Action:     log.Action,
			EntityType: log.EntityType,
			EntityID:   log.EntityID.String(),
			Summary:    log.Summary,
			Details:    log.Details,
			CreatedAt:  log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    auditLogs,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// GetDataUploadAudits handles GET /api/admin/reports/data-uploads
func (h *AuditHandler) GetDataUploadAudits(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Parse date range
	startDate, _ := parseDate(c.Query("startDate"))
	endDate, _ := parseDate(c.Query("endDate"))
	
	// Prepare input for dedicated service if available
	if h.getDataUploadAuditsService != nil {
		input := auditApp.GetDataUploadAuditsInput{
			Page:      page,
			PageSize:  pageSize,
			StartDate: startDate,
			EndDate:   endDate,
		}
		
		// Get data upload audits
		output, err := h.getDataUploadAuditsService.GetDataUploadAudits(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Failed to get data upload audits: " + err.Error(),
				Details: "An error occurred while retrieving data upload history. Please try again later.",
			})
			return
		}
		
		// Prepare response
		audits := make([]response.DataUploadAuditResponse, 0, len(output.Audits))
		for _, audit := range output.Audits {
			audits = append(audits, response.DataUploadAuditResponse{
				ID:                  audit.ID.String(),
				UploadedBy:          audit.UploadedBy.String(),
				UploadedAt:          audit.UploadedAt.Format("2006-01-02 15:04:05"),
				FileName:            audit.FileName,
				TotalUploaded:       audit.TotalUploaded,
				SuccessfullyImported: audit.SuccessfullyImported,
				DuplicatesSkipped:   audit.DuplicatesSkipped,
				ErrorsEncountered:   audit.ErrorsEncountered,
				Status:              audit.Status,
				Details:             audit.Details,
				OperationType:       audit.OperationType,
			})
		}
		
		c.JSON(http.StatusOK, response.SuccessResponse{
			Success: true,
			Data:    audits,
		})
		return
	}
	
	// Fallback to using general audit logs if dedicated service is not available
	// Prepare input - using the same GetAuditLogsInput but with specific filters for data uploads
	input := auditApp.GetAuditLogsInput{
		Page:     page,
		PageSize: pageSize,
		Filters: auditApp.AuditLogFilters{
			Action:     "UPLOAD_PARTICIPANTS", // Filter for participant upload actions
			EntityType: "Participant",
			StartDate:  startDate,
			EndDate:    endDate,
		},
	}
	
	// Get audit logs
	output, err := h.getAuditLogsService.GetAuditLogs(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get data upload audits: " + err.Error(),
			Details: "An error occurred while retrieving data upload history. Please try again later.",
		})
		return
	}
	
	// Prepare response - transform audit logs to data upload format
	dataUploads := make([]response.DataUploadAuditResponse, 0, len(output.AuditLogs))
	for _, auditLog := range output.AuditLogs {
		dataUploads = append(dataUploads, response.DataUploadAuditResponse{
			ID:                  auditLog.ID.String(),
			UploadedBy:          auditLog.UserID.String(),
			UploadedAt:          auditLog.CreatedAt.Format("2006-01-02 15:04:05"),
			TotalUploaded:       0, // This would be extracted from the audit log details in a real implementation
			SuccessfullyImported: 0,
			DuplicatesSkipped:   0,
			ErrorsEncountered:   0,
			Status:              "Completed",
			Details:             auditLog.Details,
			OperationType:       "Upload",
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    dataUploads,
	})
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", dateStr)
}
