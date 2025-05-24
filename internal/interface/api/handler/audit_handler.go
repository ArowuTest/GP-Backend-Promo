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
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// AuditHandler handles HTTP requests related to audit logs
type AuditHandler struct {
	getAuditLogsService       *auditApp.GetAuditLogsService
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(
	getAuditLogsService *auditApp.GetAuditLogsService,
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService,
) *AuditHandler {
	return &AuditHandler{
		getAuditLogsService:       getAuditLogsService,
		getDataUploadAuditsService: getDataUploadAuditsService,
	}
}

// GetAuditLogs handles GET /api/admin/audit-logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse request parameters using a struct that matches your actual DTO
	var req struct {
		Page      int    `form:"page"`
		PageSize  int    `form:"pageSize"`
		UserID    string `form:"userId"`
		Action    string `form:"action"`
		StartDate string `form:"startDate"`
		EndDate   string `form:"endDate"`
	}
	
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Parse user ID if provided
	var userID uuid.UUID
	if req.UserID != "" {
		var err error
		userID, err = uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format",
			})
			return
		}
	}
	
	// Parse dates if provided
	startDate := util.ParseTimeOrZero(req.StartDate, time.RFC3339)
	endDate := util.ParseTimeOrZero(req.EndDate, time.RFC3339)
	
	// Prepare input with nested filters structure
	input := auditApp.GetAuditLogsInput{
		Page:     req.Page,
		PageSize: req.PageSize,
		Filters: auditApp.AuditLogFilters{
			StartDate: startDate,
			EndDate:   endDate,
			UserID:    userID,
			Action:    req.Action,
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
	
	// Prepare response with explicit type conversions at DTO boundary
	auditLogs := make([]response.AuditLogResponse, 0, len(output.AuditLogs))
	for _, al := range output.AuditLogs {
		// Convert UUID to string for all ID fields
		entityIDStr := al.EntityID
		if _, err := uuid.Parse(al.EntityID); err == nil {
			// If it's a valid UUID, keep it as is
		} else if al.EntityID != "" {
			// If it's not empty and not a UUID, keep as is
		} else {
			// If empty, use empty string
			entityIDStr = ""
		}
		
		// Use Description for Summary if Details is not available
		details := ""
		if hasField(al, "Details") {
			details = al.Description // Fallback to Description if Details doesn't exist
		} else {
			details = al.Description
		}
		
		auditLogs = append(auditLogs, response.AuditLogResponse{
			ID:         al.ID.String(),
			UserID:     al.UserID.String(),
			Username:   al.Username,
			Action:     al.Action,
			EntityType: al.EntityType,
			EntityID:   entityIDStr,
			Summary:    al.Description,
			Details:    details,
			CreatedAt:  util.FormatTimeOrEmpty(al.CreatedAt, time.RFC3339),
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
	
	// Parse date range if provided
	startDateStr := c.DefaultQuery("startDate", "")
	endDateStr := c.DefaultQuery("endDate", "")
	startDate := util.ParseTimeOrZero(startDateStr, time.RFC3339)
	endDate := util.ParseTimeOrZero(endDateStr, time.RFC3339)
	
	// Prepare input
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
		})
		return
	}
	
	// Prepare response with explicit type conversions at DTO boundary
	dataUploadAudits := make([]response.DataUploadAuditResponse, 0, len(output.Audits))
	for _, dua := range output.Audits {
		dataUploadAudits = append(dataUploadAudits, response.DataUploadAuditResponse{
			ID:                   dua.ID.String(),
			UploadedBy:           dua.UploadedBy.String(),
			UploadedAt:           util.FormatTimeOrEmpty(dua.UploadedAt, time.RFC3339),
			FileName:             dua.FileName,
			TotalUploaded:        dua.TotalUploaded,
			SuccessfullyImported: dua.SuccessfullyImported,
			DuplicatesSkipped:    dua.DuplicatesSkipped,
			ErrorsEncountered:    dua.ErrorsEncountered,
			Status:               dua.Status,
			Details:              dua.Details,
			OperationType:        dua.OperationType,
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    dataUploadAudits,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// Helper function to check if a struct has a field
func hasField(obj interface{}, fieldName string) bool {
	// This is a simplified implementation
	// In a real implementation, you would use reflection to check if the field exists
	// For now, we'll just return false to use the fallback value
	return false
}
