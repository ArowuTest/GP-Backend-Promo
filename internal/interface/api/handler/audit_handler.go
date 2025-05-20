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
	// Parse request parameters
	var req request.GetAuditLogsRequest
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
	
	// Prepare input
	input := auditApp.GetAuditLogsInput{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		UserID:    userID,
		Action:    req.Action,
		Page:      req.Page,
		PageSize:  req.PageSize,
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
	for _, al := range output.AuditLogs {
		auditLogs = append(auditLogs, response.AuditLogResponse{
			ID:         al.ID.String(),
			UserID:     al.UserID.String(),
			Username:   al.Username,
			Action:     al.Action,
			EntityType: al.EntityType,
			EntityID:   al.EntityID.String(),
			Summary:    al.Summary,
			Details:    al.Details,
			CreatedAt:  al.CreatedAt.Format(time.RFC3339),
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
	
	// Prepare input
	input := auditApp.GetDataUploadAuditsInput{
		Page:     page,
		PageSize: pageSize,
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
	
	// Prepare response
	dataUploadAudits := make([]response.DataUploadAuditResponse, 0, len(output.DataUploadAudits))
	for _, dua := range output.DataUploadAudits {
		dataUploadAudits = append(dataUploadAudits, response.DataUploadAuditResponse{
			ID:                  dua.ID.String(),
			UploadedBy:          dua.UploadedBy.String(),
			UploadedAt:          dua.UploadedAt.Format(time.RFC3339),
			FileName:            dua.FileName,
			TotalUploaded:       dua.TotalUploaded,
			SuccessfullyImported: dua.SuccessfullyImported,
			DuplicatesSkipped:   dua.DuplicatesSkipped,
			ErrorsEncountered:   dua.ErrorsEncountered,
			Status:              dua.Status,
			Details:             dua.Details,
			OperationType:       dua.OperationType,
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