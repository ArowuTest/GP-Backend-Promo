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
	adminID, err := uuid.Parse(adminIDClaim.(string))
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

	// Validate header row based on the provided template: MSISDN,amount,optInStatus,Points,timestamp
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
		// amountStr := strings.TrimSpace(row[1]) // Not storing amount for now, awaiting clarification
		// optInStatusStr := strings.TrimSpace(row[2]) // Not storing optInStatus for now, awaiting clarification
		pointsStr := strings.TrimSpace(row[3])
		timestampStr := strings.TrimSpace(row[4])

		if msisdn == "" {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d: MSISDN is empty", rowCount))
			continue
		}

		points, err := strconv.Atoi(pointsStr)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid Points value ", rowCount, msisdn) + fmt.Sprintf("\"%s\"", pointsStr))
			continue
		}

		var optInDate *time.Time
		if timestampStr != "" {
			// Attempt to parse DD/MM/YYYY HH:MM
			parsedDate, err := time.Parse("02/01/2006 15:04", timestampStr) // Corrected format string
			if err != nil {
				// Attempt to parse DD/MM/YYYY if HH:MM is missing or causes error (e.g. if time is 00:00)
                parsedDateOnly, errOnly := time.Parse("02/01/2006", timestampStr)
                if errOnly != nil {
                    errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid timestamp format ", rowCount, msisdn) + fmt.Sprintf("\"%s\". Expected DD/MM/YYYY HH:MM or DD/MM/YYYY", timestampStr))
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

	auditEntry.RecordCount = rowCount -1 // Subtract header row for actual data row count

	if len(participantsToCreate) > 0 {
		tx := config.DB.Clauses(gorm.Clause.OnConflict{
			Columns:   []gorm.Column{{Name: "msisdn"}},
			DoUpdates: gorm.AssignmentColumns([]string{"points", "opt_in_date", "updated_at"}),
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
        if len(errorMessages) == 0 { // Only add this if no other errors were present
            errorMessages = append(errorMessages, "No data rows found in the CSV file after the header.")
        }
    } else { // No new participants to create, but previous rows might have had errors
        if len(errorMessages) > 0 {
            auditEntry.Status = "Partial Success"
        } else {
            auditEntry.Status = "Success" // No data to import, but file was valid and empty (or all duplicates)
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

