package interface

import (
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit/entity"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// AuditHandler handles HTTP requests related to audit logs
type AuditHandler struct {
	createAuditLog *audit.CreateAuditLogUseCase
	listAuditLogs  *audit.ListAuditLogsUseCase
	getAuditLog    *audit.GetAuditLogUseCase
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(
	createAuditLog *audit.CreateAuditLogUseCase,
	listAuditLogs *audit.ListAuditLogsUseCase,
	getAuditLog *audit.GetAuditLogUseCase,
) *AuditHandler {
	return &AuditHandler{
		createAuditLog: createAuditLog,
		listAuditLogs:  listAuditLogs,
		getAuditLog:    getAuditLog,
	}
}

// CreateAuditLog handles the request to create an audit log
func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	var req request.CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Parse user ID
	var userID *uuid.UUID
	if req.UserID != "" {
		parsedUserID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID",
				Details: err.Error(),
			})
			return
		}
		userID = &parsedUserID
	}

	// Create audit log
	input := audit.CreateAuditLogInput{
		Action:      req.Action,
		Module:      req.Module,
		Description: req.Description,
		UserID:      userID,
		Metadata:    req.Metadata,
	}

	output, err := h.createAuditLog.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create audit log",
			Details: err.Error(),
		})
		return
	}

	// Format user ID for response
	var userIDStr string
	if output.AuditLog.UserID != nil {
		userIDStr = output.AuditLog.UserID.String()
	}

	// Prepare response
	resp := response.AuditLogResponse{
		ID:          output.AuditLog.ID.String(),
		Action:      output.AuditLog.Action,
		Module:      output.AuditLog.Module,
		Description: output.AuditLog.Description,
		UserID:      userIDStr,
		Metadata:    output.AuditLog.Metadata,
		CreatedAt:   output.AuditLog.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// ListAuditLogs handles the request to list audit logs
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Parse filters
	module := c.Query("module")
	action := c.Query("action")
	
	var userID *uuid.UUID
	userIDStr := c.Query("userId")
	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID",
				Details: err.Error(),
			})
			return
		}
		userID = &parsedUserID
	}
	
	var startDate *time.Time
	startDateStr := c.Query("startDate")
	if startDateStr != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid start date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
		startDate = &parsedStartDate
	}
	
	var endDate *time.Time
	endDateStr := c.Query("endDate")
	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid end date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
		// Set to end of day
		parsedEndDate = parsedEndDate.Add(24*time.Hour - 1*time.Second)
		endDate = &parsedEndDate
	}

	// List audit logs
	input := audit.ListAuditLogsInput{
		Page:      page,
		PageSize:  pageSize,
		Module:    module,
		Action:    action,
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	output, err := h.listAuditLogs.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list audit logs",
			Details: err.Error(),
		})
		return
	}

	// Convert audit logs to response format
	auditLogs := make([]response.AuditLogResponse, 0, len(output.AuditLogs))
	for _, al := range output.AuditLogs {
		// Format user ID for response
		var userIDStr string
		if al.UserID != nil {
			userIDStr = al.UserID.String()
		}

		auditLog := response.AuditLogResponse{
			ID:          al.ID.String(),
			Action:      al.Action,
			Module:      al.Module,
			Description: al.Description,
			UserID:      userIDStr,
			Metadata:    al.Metadata,
			CreatedAt:   al.CreatedAt.Format(time.RFC3339),
		}
		auditLogs = append(auditLogs, auditLog)
	}

	// Prepare response
	resp := response.PaginatedResponse{
		Success: true,
		Data:    auditLogs,
		Pagination: response.Pagination{
			Page:      page,
			PageSize:  pageSize,
			TotalRows: output.Total,
			TotalPages: (output.Total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetAuditLog handles the request to get an audit log by ID
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	// Parse audit log ID
	auditLogIDStr := c.Param("id")
	auditLogID, err := uuid.Parse(auditLogIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid audit log ID",
			Details: err.Error(),
		})
		return
	}

	// Get audit log
	input := audit.GetAuditLogInput{
		AuditLogID: auditLogID,
	}

	output, err := h.getAuditLog.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.AuditError:
			auditErr := err.(*entity.AuditError)
			if auditErr.Code() == entity.ErrAuditLogNotFound {
				statusCode = http.StatusNotFound
				errorMessage = "Audit log not found"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to get audit log"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to get audit log"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Format user ID for response
	var userIDStr string
	if output.AuditLog.UserID != nil {
		userIDStr = output.AuditLog.UserID.String()
	}

	// Prepare response
	resp := response.AuditLogResponse{
		ID:          output.AuditLog.ID.String(),
		Action:      output.AuditLog.Action,
		Module:      output.AuditLog.Module,
		Description: output.AuditLog.Description,
		UserID:      userIDStr,
		Metadata:    output.AuditLog.Metadata,
		CreatedAt:   output.AuditLog.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}
