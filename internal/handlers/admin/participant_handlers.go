package admin

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause" // Correct import for GORM clauses
)

// HandleParticipantUpload processes the uploaded CSV file for participants.
func HandleParticipantUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request: " + err.Error()})
		return
	}
	defer file.Close()

	adminIDClaim, _ := c.Get("userID") // Assuming userID is set by auth middleware
	adminIDStr, ok := adminIDClaim.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin user ID in token is not a string"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID in token"})
		return
	}

	auditEntry := models.DataUploadAudit{
		UploadedByUserID: adminID,
		UploadTimestamp:  time.Now(),
		FileName:         header.Filename,
		OperationType:    "ParticipantUpload",
		Status:           "Pending",
	}

	if err := config.DB.Create(&auditEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit entry: " + err.Error()})
		return
	}

	reader := csv.NewReader(bufio.NewReader(file))
	var participantsToCreate []models.Participant
	var errorMessages []string
	rowCount := 0
	successfulRowCount := 0

	// Read header row
	headerRow, err := reader.Read()
	if err != nil {
		auditEntry.Status = "Failed"
		auditEntry.Notes = "Failed to read CSV header: " + err.Error()
		config.DB.Save(&auditEntry)
		c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
		return
	}
	rowCount++ // Account for header row in processed rows

	expectedHeaders := []string{"msisdn", "amount", "optinstatus", "points", "timestamp"}
	if len(headerRow) < len(expectedHeaders) {
		auditEntry.Status = "Failed"
		auditEntry.Notes = fmt.Sprintf("Invalid CSV header. Expected %d columns, got %d. Headers should be: MSISDN, amount, optInStatus, Points, timestamp", len(expectedHeaders), len(headerRow))
		config.DB.Save(&auditEntry)
		c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
		return
	}

	for i, expected := range expectedHeaders {
		if strings.ToLower(strings.TrimSpace(headerRow[i])) != expected {
			auditEntry.Status = "Failed"
			auditEntry.Notes = fmt.Sprintf("Invalid CSV header. Column %d should be ", i+1) + strings.Title(expected) + fmt.Sprintf(" (case-insensitive), but got %s", headerRow[i])
			config.DB.Save(&auditEntry)
			c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
			return
		}
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Error reading row %d: %s", rowCount+1, err.Error()))
			rowCount++
			continue
		}
		rowCount++

		if len(row) < len(expectedHeaders) {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d: Not enough columns. Expected %d, got %d.", rowCount, len(expectedHeaders), len(row)))
			continue
		}

		msisdn := strings.TrimSpace(row[0])
		pointsStr := strings.TrimSpace(row[3])
		timestampStr := strings.TrimSpace(row[4])

		if msisdn == "" {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d: MSISDN is empty", rowCount))
			continue
		}

		points, err := strconv.Atoi(pointsStr)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid Points value \"%s\"", rowCount, msisdn, pointsStr))
			continue
		}

		var optInDate *time.Time
		if timestampStr != "" {
			parsedDate, err := time.Parse("02/01/2006 15:04", timestampStr)
			if err != nil {
				parsedDateOnly, errOnly := time.Parse("02/01/2006", timestampStr)
				if errOnly != nil {
					errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid timestamp format \"%s\". Expected DD/MM/YYYY HH:MM or DD/MM/YYYY", rowCount, msisdn, timestampStr))
				} else {
					optInDate = &parsedDateOnly
				}
			} else {
				optInDate = &parsedDate
			}
		}

		participantsToCreate = append(participantsToCreate, models.Participant{
			MSISDN:    msisdn,
			Points:    points,
			OptInDate: optInDate,
		})
		successfulRowCount++
	}

	auditEntry.RecordCount = rowCount - 1 // Subtract header row for actual data row count

	if len(participantsToCreate) > 0 {
		tx := config.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "msisdn"}},
			DoUpdates: clause.AssignmentColumns([]string{"points", "opt_in_date", "updated_at"}),
		}).Create(&participantsToCreate)

		if tx.Error != nil {
			errorMessages = append(errorMessages, "Database error during bulk insert/update: "+tx.Error.Error())
			auditEntry.Status = "Failed"
		} else {
			if len(errorMessages) > 0 {
				auditEntry.Status = "Partial Success"
			} else {
				auditEntry.Status = "Success"
			}
		}
	} else if auditEntry.RecordCount > 0 && len(errorMessages) == auditEntry.RecordCount {
		auditEntry.Status = "Failed"
	} else if auditEntry.RecordCount == 0 {
		auditEntry.Status = "Failed"
		if len(errorMessages) == 0 {
			errorMessages = append(errorMessages, "No data rows found in the CSV file after the header.")
		}
	} else {
		if len(errorMessages) > 0 {
			auditEntry.Status = "Partial Success"
		} else {
			auditEntry.Status = "Success"
		}
	}

	auditEntry.Notes = strings.Join(errorMessages, "; ")
	if err := config.DB.Save(&auditEntry).Error; err != nil {
		fmt.Printf("Failed to update audit entry %s: %s\n", auditEntry.ID, err.Error())
	}

	if auditEntry.Status == "Failed" && auditEntry.RecordCount == 0 && len(participantsToCreate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "File processed. No valid participant data found to upload.",
			"audit_id": auditEntry.ID,
			"status": auditEntry.Status,
			"total_data_rows_processed": auditEntry.RecordCount,
			"successful_rows_imported": successfulRowCount,
			"errors": errorMessages,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File processed.",
		"audit_id": auditEntry.ID,
		"status": auditEntry.Status,
		"total_data_rows_processed": auditEntry.RecordCount,
		"successful_rows_imported": successfulRowCount,
		"errors": errorMessages,
	})
}

