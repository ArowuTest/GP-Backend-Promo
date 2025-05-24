package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	auditApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	getAuditLogsService       *auditApp.GetAuditLogsServiceImpl
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(
	getAuditLogsService *auditApp.GetAuditLogsServiceImpl,
	getDataUploadAuditsService *auditApp.GetDataUploadAuditsService,
) *AuditHandler {
	return &AuditHandler{
		getAuditLogsService:       getAuditLogsService,
		getDataUploadAuditsService: getDataUploadAuditsService,
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

	// Parse request parameters
	userIDStr := c.DefaultQuery("userId", "")
	action := c.DefaultQuery("action", "")
	startDateStr := c.DefaultQuery("startDate", "")
	endDateStr := c.DefaultQuery("endDate", "")

	// Parse user ID if provided
	var userID uuid.UUID
	if userIDStr != "" {
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format",
				Details: "User ID must be a valid UUID",
			})
			return
		}
	}

	// Parse dates if provided
	startDate := util.ParseTimeOrZero(startDateStr, time.RFC3339)
	endDate := util.ParseTimeOrZero(endDateStr, time.RFC3339)

	// Prepare input with flat structure
	input := auditApp.GetAuditLogsInput{
		Page:      page,
		PageSize:  pageSize,
		StartDate: &startDate,
		EndDate:   &endDate,
		PerformedBy: &userID,
		Action:    action,
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
		entityIDStr := al.EntityID.String()
		
		// Extract details from metadata if available
		details := ""
		if al.Metadata != nil {
			if detailsVal, ok := al.Metadata["details"]; ok {
				if detailsStr, ok := detailsVal.(string); ok {
					details = detailsStr
				}
			}
		}

		auditLogs = append(auditLogs, response.AuditLogResponse{
			ID:         al.ID.String(),
			UserID:     al.PerformedBy.String(),
			Username:   "", // Not available in output
			Action:     al.Action,
			EntityType: al.Entity,
			EntityID:   entityIDStr,
			Summary:    "", // Not available in output
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
