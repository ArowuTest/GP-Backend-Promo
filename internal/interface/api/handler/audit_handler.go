package handler

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	auditApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	getAuditLogsService *auditApp.GetAuditLogsService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(getAuditLogsService *auditApp.GetAuditLogsService) *AuditHandler {
	return &AuditHandler{
		getAuditLogsService: getAuditLogsService,
	}
}

// GetAuditLogs handles GET /api/admin/audit-logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
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
		})
		return
	}
	
	// Prepare response
	auditLogs := make([]response.AuditLogResponse, 0, len(output.AuditLogs))
	for _, auditLog := range output.AuditLogs {
		auditLogs = append(auditLogs, response.AuditLogResponse{
			ID:         auditLog.ID.String(),
			Action:     auditLog.Action,
			EntityType: auditLog.EntityType,
			EntityID:   auditLog.EntityID,
			UserID:     auditLog.UserID.String(),
			Summary:    auditLog.Description,
			Details:    auditLog.Description,
			CreatedAt:  auditLog.CreatedAt.Format("2006-01-02 15:04:05"),
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
		})
		return
	}
	
	// Prepare response - transform audit logs to data upload format
	dataUploads := make([]response.DataUploadAuditResponse, 0, len(output.AuditLogs))
	for _, auditLog := range output.AuditLogs {
		dataUploads = append(dataUploads, response.DataUploadAuditResponse{
			ID:            auditLog.ID.String(),
			UploadedBy:    auditLog.UserID.String(),
			UploadedAt:    auditLog.CreatedAt.Format("2006-01-02 15:04:05"),
			TotalUploaded: 0, // This would be extracted from the audit log details in a real implementation
			Status:        "Completed",
			Details:       auditLog.Description,
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    dataUploads,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", dateStr)
}
