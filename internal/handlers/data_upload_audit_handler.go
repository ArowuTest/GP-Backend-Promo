package handlers

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	// "strconv" // May need for pagination or filtering
)

// CreateDataUploadAuditEntry godoc
// @Summary Create a new data upload audit entry
// @Description Logs an event related to data upload or manipulation.
// @Tags DataUploadAudits
// @Accept json
// @Produce json
// @Param audit_entry body models.DataUploadAudit true "Data Upload Audit Entry object to be created"
// @Success 201 {object} models.DataUploadAudit
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/audits/data-uploads [post]
func CreateDataUploadAuditEntry(c *gin.Context) {
	var newAuditEntry models.DataUploadAudit
	if err := c.ShouldBindJSON(&newAuditEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation
	if newAuditEntry.UploadedBy == "" || newAuditEntry.OperationType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UploadedBy and OperationType are required fields"})
		return
	}

	// Set timestamp if not provided (though usually it should be set by the caller)
	if newAuditEntry.UploadedAt.IsZero() {
		newAuditEntry.UploadedAt = time.Now()
	}

	if err := db.DB.Create(&newAuditEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create data upload audit entry: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newAuditEntry)
}

// ListDataUploadAuditEntries godoc
// @Summary List all data upload audit entries
// @Description Get a list of all data upload audit entries, potentially with pagination.
// @Tags DataUploadAudits
// @Produce json
// @Success 200 {array} models.DataUploadAudit
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/audits/data-uploads [get]
func ListDataUploadAuditEntries(c *gin.Context) {
	var auditEntries []models.DataUploadAudit
	// Add pagination later if needed: e.g., c.Query("page"), c.Query("limit")
	if err := db.DB.Order("uploaded_at desc").Find(&auditEntries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data upload audit entries: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, auditEntries)
}

// Potentially add GetDataUploadAuditEntryByID if needed later

