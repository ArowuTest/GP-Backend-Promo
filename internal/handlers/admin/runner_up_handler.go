package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RunnerUpHandler handles operations related to runner-ups
type RunnerUpHandler struct {
	db             *gorm.DB
	drawDataService services.DrawDataService
	auditService    *services.AuditService
}

// NewRunnerUpHandler creates a new RunnerUpHandler
func NewRunnerUpHandler(db *gorm.DB, drawDataService services.DrawDataService, auditService *services.AuditService) *RunnerUpHandler {
	return &RunnerUpHandler{
		db:             db,
		drawDataService: drawDataService,
		auditService:    auditService,
	}
}

// ListRunnerUps handles listing all runner-ups
func (h *RunnerUpHandler) ListRunnerUps(c *gin.Context) {
	// Parse query parameters
	drawIDStr := c.Query("draw_id")
	prizeIDStr := c.Query("prize_id")
	status := c.Query("status")

	// Build query
	query := h.db.Model(&models.RunnerUp{}).
		Preload("Draw").
		Preload("Prize")

	// Apply filters
	if drawIDStr != "" {
		drawID, err := uuid.Parse(drawIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw ID format"})
			return
		}
		query = query.Where("draw_id = ?", drawID)
	}

	if prizeIDStr != "" {
		prizeID, err := uuid.Parse(prizeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize ID format"})
			return
		}
		query = query.Where("prize_id = ?", prizeID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get runner-ups
	var runnerUps []models.RunnerUp
	if err := query.Find(&runnerUps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner-ups: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, runnerUps)
}

// GetRunnerUp handles retrieving a single runner-up by ID
func (h *RunnerUpHandler) GetRunnerUp(c *gin.Context) {
	runnerUpIDStr := c.Param("id")
	runnerUpID, err := uuid.Parse(runnerUpIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner-up ID format"})
		return
	}

	var runnerUp models.RunnerUp
	if err := h.db.Preload("Draw").
		Preload("Prize").
		First(&runnerUp, runnerUpID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner-up not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner-up: " + err.Error()})
		}
		return
	}

	// Return response
	c.JSON(http.StatusOK, runnerUp)
}

// UpdateRunnerUpStatus handles updating the status of a runner-up
func (h *RunnerUpHandler) UpdateRunnerUpStatus(c *gin.Context) {
	runnerUpIDStr := c.Param("id")
	runnerUpID, err := uuid.Parse(runnerUpIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner-up ID format"})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"notified":  true,
		"promoted":  true,
		"skipped":   true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Valid values: pending, notified, promoted, skipped"})
		return
	}

	// Get admin ID from context
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	// Get runner-up record
	var runnerUp models.RunnerUp
	if err := h.db.First(&runnerUp, runnerUpID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner-up not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner-up: " + err.Error()})
		}
		return
	}

	// Update status
	updates := map[string]interface{}{
		"status": req.Status,
	}

	// Add status-specific fields
	switch req.Status {
	case "notified":
		updates["notified_at"] = time.Now()
		updates["notified_by_admin_id"] = adminID
	case "promoted":
		updates["promoted_at"] = time.Now()
		updates["promoted_by_admin_id"] = adminID
	case "skipped":
		updates["skipped_at"] = time.Now()
		updates["skipped_by_admin_id"] = adminID
	}

	if err := h.db.Model(&runnerUp).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner-up status: " + err.Error()})
		return
	}

	// Log audit event
	auditEvent := models.AuditLog{
		AdminID:     adminID,
		Action:      "update_runner_up_status",
		EntityType:  "runner_up",
		EntityID:    runnerUp.ID.String(),
		Description: fmt.Sprintf("Updated runner-up status to %s for MSISDN %s", req.Status, runnerUp.MSISDN),
	}
	h.auditService.LogAuditEvent(auditEvent)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Runner-up status updated successfully",
		"runner_up": runnerUp,
	})
}
