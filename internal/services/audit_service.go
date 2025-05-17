package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditService provides methods for system-wide audit logging
type AuditService struct{}

// NewAuditService creates a new AuditService
func NewAuditService() *AuditService {
	return &AuditService{}
}

// LogUserAction logs a user action to the system audit log
func (s *AuditService) LogUserAction(c *gin.Context, actionType, resourceType, resourceID, description string, actionDetails interface{}) error {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		return fmt.Errorf("user ID not found in context")
	}

	// Convert actionDetails to JSON string
	var actionDetailsStr string
	if actionDetails != nil {
		detailsBytes, err := json.Marshal(actionDetails)
		if err != nil {
			return fmt.Errorf("failed to marshal action details: %w", err)
		}
		actionDetailsStr = string(detailsBytes)
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create audit log entry
	err := models.CreateSystemAuditLog(
		config.DB,
		userID.(uuid.UUID),
		actionType,
		resourceType,
		resourceID,
		description,
		ipAddress,
		userAgent,
		actionDetailsStr,
	)

	return err
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(
	startDate, endDate *time.Time,
	userID *uuid.UUID,
	actionType, resourceType, resourceID *string,
	page, pageSize int,
) ([]models.SystemAuditLog, int64, error) {
	query := config.DB.Model(&models.SystemAuditLog{}).Preload("User")

	// Apply filters
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}
	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if actionType != nil && *actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if resourceType != nil && *resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if resourceID != nil && *resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply pagination
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Order by created_at desc
	query = query.Order("created_at DESC")

	// Execute query
	var auditLogs []models.SystemAuditLog
	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// Middleware to automatically log API requests
func (s *AuditService) AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for certain endpoints like health checks or public endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/v1/auth/login" {
			c.Next()
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			// If no user ID, this might be a public endpoint or auth failed
			c.Next()
			return
		}

		// Prepare audit log data
		actionType := "API_REQUEST"
		resourceType := "ENDPOINT"
		resourceID := c.Request.URL.Path
		description := fmt.Sprintf("%s request to %s", c.Request.Method, c.Request.URL.Path)

		// Capture request details
		requestDetails := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"query":       c.Request.URL.Query(),
			"contentType": c.ContentType(),
			"statusCode":  c.Writer.Status(),
		}

		// Execute the request
		c.Next()

		// Update with response status
		requestDetails["statusCode"] = c.Writer.Status()
		requestDetails["responseSize"] = c.Writer.Size()

		// Only log successful requests (status 2xx) or error requests (status 4xx/5xx)
		statusCode := c.Writer.Status()
		if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices ||
			statusCode >= http.StatusBadRequest {
			
			// Convert requestDetails to JSON
			detailsBytes, err := json.Marshal(requestDetails)
			if err != nil {
				// Just log the error but don't interrupt the request flow
				fmt.Printf("Failed to marshal request details for audit log: %v\n", err)
				return
			}

			// Create audit log entry
			err = models.CreateSystemAuditLog(
				config.DB,
				userID.(uuid.UUID),
				actionType,
				resourceType,
				resourceID,
				description,
				c.ClientIP(),
				c.Request.UserAgent(),
				string(detailsBytes),
			)

			if err != nil {
				// Just log the error but don't interrupt the request flow
				fmt.Printf("Failed to create audit log: %v\n", err)
			}
		}
	}
}
