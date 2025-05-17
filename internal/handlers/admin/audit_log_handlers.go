package admin

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditLogHandler handles requests related to system audit logs
type AuditLogHandler struct {
	AuditService *services.AuditService
}

// NewAuditLogHandler creates a new AuditLogHandler
func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{
		AuditService: services.NewAuditService(),
	}
}

// ListSystemAuditLogs godoc
// @Summary List system audit logs
// @Description Retrieves a paginated list of system audit logs with optional filtering
// @Tags Admin,Audit
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param user_id query string false "Filter by user ID"
// @Param action_type query string false "Filter by action type"
// @Param resource_type query string false "Filter by resource type"
// @Param resource_id query string false "Filter by resource ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/audit-logs [get]
func (h *AuditLogHandler) ListSystemAuditLogs(c *gin.Context) {
	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userIDStr := c.Query("user_id")
	actionType := c.Query("action_type")
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	
	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if _, err := c.GetQuery("page"); err {
			page = 1
		}
	}
	
	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if _, err := c.GetQuery("page_size"); err {
			pageSize = 20
		}
	}
	
	// Limit max page size
	if pageSize > 100 {
		pageSize = 100
	}
	
	// Parse date parameters
	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Expected YYYY-MM-DD"})
			return
		}
		startDate = &parsedDate
	}
	
	if endDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Expected YYYY-MM-DD"})
			return
		}
		// Set to end of day
		parsedDate = parsedDate.Add(24*time.Hour - time.Second)
		endDate = &parsedDate
	}
	
	// Parse user ID
	var userID *uuid.UUID
	if userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		userID = &parsedID
	}
	
	// Set action_type pointer
	var actionTypePtr *string
	if actionType != "" {
		actionTypePtr = &actionType
	}
	
	// Set resource_type pointer
	var resourceTypePtr *string
	if resourceType != "" {
		resourceTypePtr = &resourceType
	}
	
	// Set resource_id pointer
	var resourceIDPtr *string
	if resourceID != "" {
		resourceIDPtr = &resourceID
	}
	
	// Get audit logs
	auditLogs, total, err := h.AuditService.GetAuditLogs(
		startDate,
		endDate,
		userID,
		actionTypePtr,
		resourceTypePtr,
		resourceIDPtr,
		page,
		pageSize,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs: " + err.Error()})
		return
	}
	
	// Calculate pagination metadata
	totalPages := (int(total) + pageSize - 1) / pageSize
	hasNextPage := page < totalPages
	hasPrevPage := page > 1
	
	// Return response
	c.JSON(http.StatusOK, gin.H{
		"data": auditLogs,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"has_next":    hasNextPage,
			"has_prev":    hasPrevPage,
		},
	})
}

// GetAuditLogTypes godoc
// @Summary Get available audit log types
// @Description Retrieves lists of available action types and resource types for filtering
// @Tags Admin,Audit
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/audit-logs/types [get]
func (h *AuditLogHandler) GetAuditLogTypes(c *gin.Context) {
	// Get distinct action types
	var actionTypes []string
	if err := config.DB.Model(&models.SystemAuditLog{}).Distinct().Pluck("action_type", &actionTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch action types: " + err.Error()})
		return
	}
	
	// Get distinct resource types
	var resourceTypes []string
	if err := config.DB.Model(&models.SystemAuditLog{}).Distinct().Pluck("resource_type", &resourceTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch resource types: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"action_types":   actionTypes,
		"resource_types": resourceTypes,
	})
}
