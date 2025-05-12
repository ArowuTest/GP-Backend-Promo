package handlers

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config" // Corrected to use config.DB
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // For UploadedByUserID
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
	var input struct {
		UploadedByUserID uuid.UUID `json:"uploaded_by_user_id" binding:"required"`
		FileName         string    `json:"file_name,omitempty"`
		RecordCount      int       `json:"record_count,omitempty"`
		Status           string    `json:"status,omitempty"`
		Notes            string    `json:"notes,omitempty"`
		OperationType    string    `json:"operation_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	newAuditEntry.UploadedByUserID = input.UploadedByUserID
	newAuditEntry.FileName = input.FileName
	newAuditEntry.RecordCount = input.RecordCount
	newAuditEntry.Status = input.Status
	newAuditEntry.Notes = input.Notes
	newAuditEntry.OperationType = input.OperationType
	newAuditEntry.UploadTimestamp = time.Now() // Set timestamp upon creation

	if err := config.DB.Create(&newAuditEntry).Error; err != nil { // Corrected to use config.DB
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
	if err := config.DB.Order("upload_timestamp desc").Find(&auditEntries).Error; err != nil { // Corrected to use config.DB
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data upload audit entries: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, auditEntries)
}

