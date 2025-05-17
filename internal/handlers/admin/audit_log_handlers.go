package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogHandler handles operations related to audit logs
type AuditLogHandler struct {
	db           *gorm.DB
	AuditService *services.AuditService
}

// NewAuditLogHandler creates a new AuditLogHandler
func NewAuditLogHandler(db *gorm.DB, auditService *services.AuditService) *AuditLogHandler {
	return &AuditLogHandler{
		db:           db,
		AuditService: auditService,
	}
}

// ListAuditLogs handles listing audit logs with filtering and pagination
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filter parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userIDStr := c.Query("user_id")
	actionType := c.Query("action_type")
	resourceType := c.Query("resource_type")

	// Convert string parameters to appropriate types
	var startDate, endDate *time.Time
	var userID *uuid.UUID
	var actionTypePtr, resourceTypePtr *string

	if startDateStr != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
		startDate = &parsedStartDate
	} else {
		// Default to 30 days ago if not provided
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		startDate = &defaultStartDate
	}

	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
		// Set end date to end of day
		parsedEndDate = parsedEndDate.Add(24*time.Hour - 1*time.Second)
		endDate = &parsedEndDate
	} else {
		// Default to now if not provided
		defaultEndDate := time.Now()
		endDate = &defaultEndDate
	}

	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		userID = &parsedUserID
	}

	if actionType != "" {
		actionTypePtr = &actionType
	}

	if resourceType != "" {
		resourceTypePtr = &resourceType
	}

	// Convert time.Time to Unix timestamps for the service call
	fromUnix := int(startDate.Unix())
	toUnix := int(endDate.Unix())
	
	// Convert resourceTypePtr to string (dereferencing if not nil)
	resType := ""
	if resourceTypePtr != nil {
		resType = *resourceTypePtr
	}

	// Get audit logs from service
	logs, err := h.AuditService.GetAuditLogs(fromUnix, toUnix, userID, actionTypePtr, resType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs: " + err.Error()})
		return
	}

	// Apply pagination
	totalCount := len(logs)
	offset := (page - 1) * limit
	endIndex := offset + limit
	if endIndex > totalCount {
		endIndex = totalCount
	}

	var paginatedLogs []models.AuditLog
	if offset < totalCount {
		paginatedLogs = logs[offset:endIndex]
	} else {
		paginatedLogs = []models.AuditLog{}
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"logs":        paginatedLogs,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": (totalCount + limit - 1) / limit,
	})
}

// GetAuditLog handles retrieving a single audit log by ID
func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	logIDStr := c.Param("id")
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit log ID format"})
		return
	}

	log, err := h.AuditService.GetAuditLogByID(logID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit log: " + err.Error()})
		}
		return
	}

	// Return response
	c.JSON(http.StatusOK, log)
}

// ExportAuditLogs handles exporting audit logs to CSV or JSON
func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	// Parse query parameters
	format := c.DefaultQuery("format", "json")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userIDStr := c.Query("user_id")
	actionType := c.Query("action_type")
	resourceType := c.Query("resource_type")

	// Convert string parameters to appropriate types
	var startDate, endDate *time.Time
	var userID *uuid.UUID
	var actionTypePtr, resourceTypePtr *string

	if startDateStr != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
		startDate = &parsedStartDate
	} else {
		// Default to 30 days ago if not provided
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		startDate = &defaultStartDate
	}

	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
		// Set end date to end of day
		parsedEndDate = parsedEndDate.Add(24*time.Hour - 1*time.Second)
		endDate = &parsedEndDate
	} else {
		// Default to now if not provided
		defaultEndDate := time.Now()
		endDate = &defaultEndDate
	}

	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		userID = &parsedUserID
	}

	if actionType != "" {
		actionTypePtr = &actionType
	}

	if resourceType != "" {
		resourceTypePtr = &resourceType
	}

	// Convert time.Time to Unix timestamps for the service call
	fromUnix := int(startDate.Unix())
	toUnix := int(endDate.Unix())
	
	// Convert resourceTypePtr to string (dereferencing if not nil)
	resType := ""
	if resourceTypePtr != nil {
		resType = *resourceTypePtr
	}

	// Get audit logs from service
	logs, err := h.AuditService.GetAuditLogs(fromUnix, toUnix, userID, actionTypePtr, resType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs: " + err.Error()})
		return
	}

	// Export based on format
	switch format {
	case "json":
		// Return JSON response
		c.JSON(http.StatusOK, logs)
	case "csv":
		// Generate CSV content
		csvContent := "ID,Admin ID,Admin Username,Action,Entity Type,Entity ID,Description,Created At\n"
		for _, log := range logs {
			csvContent += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
				log.ID,
				log.AdminID,
				log.AdminUsername,
				log.Action,
				log.EntityType,
				log.EntityID,
				log.Description,
				log.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}

		// Set headers for CSV download
		c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
		c.Data(http.StatusOK, "text/csv", []byte(csvContent))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export format. Supported formats: json, csv"})
	}
}
