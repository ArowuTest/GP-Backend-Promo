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

// normalizeMSISDN ensures MSISDN is in a standard format if needed (e.g., remove leading +, spaces)
func normalizeMSISDN(msisdn string) string {
	return strings.TrimSpace(msisdn) // Basic trimming, can be expanded
}

// HandleParticipantUpload processes the uploaded CSV file for participant events.
func HandleParticipantUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request: " + err.Error()})
		return
	}
	defer file.Close()

	adminIDClaim, _ := c.Get("userID")
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
	var errorMessages []string
	var skippedDuplicateDetails []string
	csvRowCount := 0
	successfullyImportedCount := 0
	duplicatesSkippedCount := 0

	headerRow, err := reader.Read()
	if err != nil {
		auditEntry.Status = "Failed"
		auditEntry.Notes = "Failed to read CSV header: " + err.Error()
		config.DB.Save(&auditEntry)
		c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
		return
	}
	csvRowCount++

	// Define expected headers (case-insensitive check)
	// msisdn, amount, optinstatus, points, timestamp
	colMap := make(map[string]int)
	expectedHeaders := []string{"msisdn", "amount", "optinstatus", "points", "timestamp"}
	for i, h := range headerRow {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, eh := range expectedHeaders {
		if _, exists := colMap[eh]; !exists {
			auditEntry.Status = "Failed"
			auditEntry.Notes = fmt.Sprintf("Missing expected header column: %s. Expected headers are: msisdn, amount, optinstatus, points, timestamp (case-insensitive).", eh)
			config.DB.Save(&auditEntry)
			c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
			return
		}
	}

	var participantEventsToCreate []models.ParticipantEvent
	var participantsToUpsert []models.Participant // For master participant records
	// Use a map to track unique events within this CSV to avoid processing exact same row multiple times if it appears in CSV
	// Key: msisdn_amount_points_timestamp_optinstatus (a composite key for uniqueness of an event)
	processedEventsInCSV := make(map[string]bool)
	// Use a map to track participants for upserting into Participant master table
	participantMasterMap := make(map[string]models.Participant)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		csvRowCount++
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Error reading CSV row %d: %s", csvRowCount, err.Error()))
			continue
		}

		msisdn := normalizeMSISDN(row[colMap["msisdn"]])
		amountStr := strings.TrimSpace(row[colMap["amount"]])
		optInStatusStr := strings.TrimSpace(row[colMap["optinstatus"]])
		pointsStr := strings.TrimSpace(row[colMap["points"]])
		timestampStr := strings.TrimSpace(row[colMap["timestamp"]])

		if msisdn == "" {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d: MSISDN is empty.", csvRowCount))
			continue
		}

		pointsEarned, err := strconv.Atoi(pointsStr)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid Points value \"%s\". Must be a number.", csvRowCount, msisdn, pointsStr))
			continue
		}

		var transactionTime *time.Time
		if timestampStr != "" {
			t, errParse := time.Parse("02/01/2006 15:04", timestampStr) // DD/MM/YYYY HH:MM
			if errParse != nil {
				tDateOnly, errParseDateOnly := time.Parse("02/01/2006", timestampStr) // DD/MM/YYYY
				if errParseDateOnly != nil {
					errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid timestamp format \"%s\". Expected DD/MM/YYYY HH:MM or DD/MM/YYYY.", csvRowCount, msisdn, timestampStr))
					continue
				}
				transactionTime = &tDateOnly
			} else {
				transactionTime = &t
			}
		} else {
		    errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Timestamp is required for participant event.", csvRowCount, msisdn))
		    continue
		}

		// Create a unique key for the event from the CSV to handle duplicates within the same file
		eventKey := fmt.Sprintf("%s_%s_%d_%s_%s", msisdn, amountStr, pointsEarned, transactionTime.Format(time.RFC3339Nano), optInStatusStr)
		if processedEventsInCSV[eventKey] {
			duplicatesSkippedCount++
			skippedDuplicateDetails = append(skippedDuplicateDetails, fmt.Sprintf("Row %d (MSISDN %s): Duplicate event within CSV file.", csvRowCount, msisdn))
			continue
		}
		processedEventsInCSV[eventKey] = true

		// Check for existing event in DB (true duplicate)
		var existingEvent models.ParticipantEvent
		result := config.DB.Where("msisdn = ? AND amount = ? AND points_earned = ? AND transaction_timestamp = ? AND opt_in_status = ?",
			msisdn, amountStr, pointsEarned, transactionTime, optInStatusStr).First(&existingEvent)

		if result.Error == nil { // Found an existing identical event in DB
			duplicatesSkippedCount++
			skippedDuplicateDetails = append(skippedDuplicateDetails, fmt.Sprintf("Row %d (MSISDN %s): Exact duplicate event already exists in database (ID: %s).", csvRowCount, msisdn, existingEvent.ID.String()))
			continue
		} else if result.Error != gorm.ErrRecordNotFound {
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): DB error checking for duplicate event: %s", csvRowCount, msisdn, result.Error.Error()))
			continue
		}

		// If we reach here, it's a new, unique event to be added
		participantEventsToCreate = append(participantEventsToCreate, models.ParticipantEvent{
			MSISDN:               msisdn,
			Amount:               amountStr,
			OptInStatus:          optInStatusStr,
			PointsEarned:         pointsEarned,
			TransactionTimestamp: transactionTime,
			UploadAuditID:        auditEntry.ID,
		})

		// Prepare participant master record for upsert
		if pMaster, exists := participantMasterMap[msisdn]; !exists {
			participantMasterMap[msisdn] = models.Participant{MSISDN: msisdn, OptInDate: transactionTime} // First time seeing this MSISDN in CSV
		} else {
		    // Update OptInDate if current event's timestamp is earlier (or based on defined logic)
		    if pMaster.OptInDate == nil || (transactionTime != nil && transactionTime.Before(*pMaster.OptInDate)) {
		        pMaster.OptInDate = transactionTime
                participantMasterMap[msisdn] = pMaster
		    }
		}
	}

	for _, p := range participantMasterMap {
	    participantsToUpsert = append(participantsToUpsert, p)
	}

	auditEntry.RecordCount = csvRowCount - 1 // Total data rows in CSV (excluding header)

	// Database transaction
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Upsert Participant master records
		if len(participantsToUpsert) > 0 {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "msisdn"}}, DoUpdates: clause.AssignmentColumns([]string{"opt_in_date", "updated_at"})}).Create(&participantsToUpsert).Error; err != nil {
				return fmt.Errorf("failed to upsert participant master records: %w", err)
			}
		}

		// 2. Create Participant Events
		if len(participantEventsToCreate) > 0 {
			if err := tx.CreateInBatches(&participantEventsToCreate, 100).Error; err != nil { // Batch insert for efficiency
				return fmt.Errorf("failed to create participant events: %w", err)
			}
			successfullyImportedCount = len(participantEventsToCreate)
		}
		return nil
	})

	if err != nil {
		errorMessages = append(errorMessages, "Database transaction error: "+err.Error())
		auditEntry.Status = "Failed"
	} else {
		if len(errorMessages) > 0 {
			auditEntry.Status = "Partial Success"
		} else if successfullyImportedCount == 0 && auditEntry.RecordCount > 0 && duplicatesSkippedCount == auditEntry.RecordCount {
            auditEntry.Status = "Success" // All records were duplicates, which is a form of success
        } else if successfullyImportedCount == 0 && auditEntry.RecordCount == 0 {
            auditEntry.Status = "Failed"
            errorMessages = append(errorMessages, "No data rows found or processed from CSV.")
        } else {
			auditEntry.Status = "Success"
		}
	}

	auditEntry.SuccessfullyImported = successfullyImportedCount
	auditEntry.DuplicatesSkipped = duplicatesSkippedCount
	notes := strings.Join(errorMessages, "; ")
	if len(skippedDuplicateDetails) > 0 {
		notes += "; Skipped Duplicates: " + strings.Join(skippedDuplicateDetails, ", ")
	}
    auditEntry.Notes = notes

	if errSaveAudit := config.DB.Save(&auditEntry).Error; errSaveAudit != nil {
		fmt.Printf("Critical: Failed to update final audit entry %s: %s\n", auditEntry.ID, errSaveAudit.Error())
		// Log this critical failure, but the main response to user might have already been determined
	}

	responseStatus := http.StatusOK
	if auditEntry.Status == "Failed" {
		responseStatus = http.StatusBadRequest
	} else if auditEntry.Status == "Partial Success" {
	    responseStatus = http.StatusMultiStatus
	}

	c.JSON(responseStatus, gin.H{
		"message":                   "File processing complete.",
		"audit_id":                  auditEntry.ID,
		"status":                    auditEntry.Status,
		"total_data_rows_in_csv":  auditEntry.RecordCount,
		"new_events_imported":       auditEntry.SuccessfullyImported,
		"true_duplicates_skipped": auditEntry.DuplicatesSkipped,
		"processing_errors":         errorMessages,
        "skipped_duplicate_details": skippedDuplicateDetails, // For more detailed client-side display if needed
	})
}

