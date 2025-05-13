package admin

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// normalizeMSISDN ensures MSISDN is in a standard format if needed (e.g., remove leading +, spaces)
func normalizeMSISDN(msisdn string) string {
	return strings.TrimSpace(msisdn) // Basic trimming, can be expanded
}

// HandleParticipantUpload processes the uploaded CSV file for participant events.
func HandleParticipantUpload(c *gin.Context) {
	log.Println("DEBUG: HandleParticipantUpload started")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("DEBUG: Error getting file from request: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request: " + err.Error()})
		return
	}
	defer file.Close()

	adminIDClaim, _ := c.Get("userID")
	adminIDStr, ok := adminIDClaim.(string)
	if !ok {
		log.Println("DEBUG: Admin user ID in token is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin user ID in token is not a string"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		log.Printf("DEBUG: Invalid admin user ID in token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID in token"})
		return
	}
	log.Printf("DEBUG: Admin User ID: %s, Uploading file: %s\n", adminID.String(), header.Filename)

	auditEntry := models.DataUploadAudit{
		UploadedByUserID: adminID,
		UploadTimestamp:  time.Now(),
		FileName:         header.Filename,
		OperationType:    "ParticipantUpload",
		Status:           "Pending",
	}

	if err := config.DB.Create(&auditEntry).Error; err != nil {
		log.Printf("DEBUG: Failed to create initial audit entry: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit entry: " + err.Error()})
		return
	}
	log.Printf("DEBUG: Initial audit entry created with ID: %s\n", auditEntry.ID.String())

	reader := csv.NewReader(bufio.NewReader(file))
	var errorMessages []string
	var skippedDuplicateDetails []string
	csvRowCount := 0
	successfullyImportedCount := 0
	duplicatesSkippedCount := 0

	headerRow, err := reader.Read()
	if err != nil {
		log.Printf("DEBUG: Failed to read CSV header: %v\n", err)
		auditEntry.Status = "Failed"
		auditEntry.Notes = "Failed to read CSV header: " + err.Error()
		config.DB.Save(&auditEntry)
		c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
		return
	}
	csvRowCount++
	log.Printf("DEBUG: CSV Header: %v\n", headerRow)

	colMap := make(map[string]int)
	expectedHeaders := []string{"msisdn", "amount", "optinstatus", "points", "timestamp"}
	for i, h := range headerRow {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, eh := range expectedHeaders {
		if _, exists := colMap[eh]; !exists {
			log.Printf("DEBUG: Missing expected header column: %s\n", eh)
			auditEntry.Status = "Failed"
			auditEntry.Notes = fmt.Sprintf("Missing expected header column: %s. Expected headers are: msisdn, amount, optinstatus, points, timestamp (case-insensitive).", eh)
			config.DB.Save(&auditEntry)
			c.JSON(http.StatusBadRequest, gin.H{"error": auditEntry.Notes})
			return
		}
	}

	var participantEventsToCreate []models.ParticipantEvent
	var participantsToUpsert []models.Participant
	processedEventsInCSV := make(map[string]bool)
	participantMasterMap := make(map[string]models.Participant)

	log.Println("DEBUG: Starting CSV row processing loop")
	for {
		row, err := reader.Read()
		if err == io.EOF {
			log.Println("DEBUG: Reached EOF of CSV file")
			break
		}
		csvRowCount++
		if err != nil {
			log.Printf("DEBUG: Error reading CSV row %d: %v\n", csvRowCount, err)
			errorMessages = append(errorMessages, fmt.Sprintf("Error reading CSV row %d: %s", csvRowCount, err.Error()))
			continue
		}
		log.Printf("DEBUG: Processing CSV Row %d: %v\n", csvRowCount, row)

		msisdn := normalizeMSISDN(row[colMap["msisdn"]])
		amountStr := strings.TrimSpace(row[colMap["amount"]])
		optInStatusStr := strings.TrimSpace(row[colMap["optinstatus"]])
		pointsStr := strings.TrimSpace(row[colMap["points"]])
		timestampStr := strings.TrimSpace(row[colMap["timestamp"]])

		if msisdn == "" {
			log.Printf("DEBUG: Row %d: MSISDN is empty. Skipping.\n", csvRowCount)
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d: MSISDN is empty.", csvRowCount))
			continue
		}

		pointsEarned, err := strconv.Atoi(pointsStr)
		if err != nil {
			log.Printf("DEBUG: Row %d (MSISDN %s): Invalid Points value \"%s\". Skipping. Error: %v\n", csvRowCount, msisdn, pointsStr, err)
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid Points value \"%s\". Must be a number.", csvRowCount, msisdn, pointsStr))
			continue
		}

		var transactionTime *time.Time
		if timestampStr != "" {
			t, errParse := time.Parse("02/01/2006 15:04", timestampStr) // DD/MM/YYYY HH:MM
			if errParse != nil {
				tDateOnly, errParseDateOnly := time.Parse("02/01/2006", timestampStr) // DD/MM/YYYY
				if errParseDateOnly != nil {
					log.Printf("DEBUG: Row %d (MSISDN %s): Invalid timestamp format \"%s\". Skipping. Error: %v\n", csvRowCount, msisdn, timestampStr, errParseDateOnly)
					errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Invalid timestamp format \"%s\". Expected DD/MM/YYYY HH:MM or DD/MM/YYYY.", csvRowCount, msisdn, timestampStr))
					continue
				}
				transactionTime = &tDateOnly
			} else {
				transactionTime = &t
			}
		} else {
			log.Printf("DEBUG: Row %d (MSISDN %s): Timestamp is required. Skipping.\n", csvRowCount, msisdn)
		    errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): Timestamp is required for participant event.", csvRowCount, msisdn))
		    continue
		}
		log.Printf("DEBUG: Row %d (MSISDN %s): Parsed data: Amount=\"%s\", OptIn=\"%s\", Points=%d, Timestamp=%v\n", csvRowCount, msisdn, amountStr, optInStatusStr, pointsEarned, transactionTime)

		eventKey := fmt.Sprintf("%s_%s_%d_%s_%s", msisdn, amountStr, pointsEarned, transactionTime.Format(time.RFC3339Nano), optInStatusStr)
		if processedEventsInCSV[eventKey] {
			log.Printf("DEBUG: Row %d (MSISDN %s): Duplicate event within this CSV file (key: %s). Skipping.\n", csvRowCount, msisdn, eventKey)
			duplicatesSkippedCount++
			skippedDuplicateDetails = append(skippedDuplicateDetails, fmt.Sprintf("Row %d (MSISDN %s): Duplicate event within CSV file.", csvRowCount, msisdn))
			continue
		}
		processedEventsInCSV[eventKey] = true

		var existingEvent models.ParticipantEvent
		log.Printf("DEBUG: Row %d (MSISDN %s): Checking DB for duplicate event. Query params: msisdn=%s, amount=%s, points=%d, timestamp=%v, optin=%s\n", csvRowCount, msisdn, msisdn, amountStr, pointsEarned, transactionTime, optInStatusStr)
		result := config.DB.Where("msisdn = ? AND amount = ? AND points_earned = ? AND transaction_timestamp = ? AND opt_in_status = ?",
			msisdn, amountStr, pointsEarned, transactionTime, optInStatusStr).First(&existingEvent)

		if result.Error == nil { 
			log.Printf("DEBUG: Row %d (MSISDN %s): Exact duplicate event already exists in database (ID: %s). Skipping.\n", csvRowCount, msisdn, existingEvent.ID.String())
			duplicatesSkippedCount++
			skippedDuplicateDetails = append(skippedDuplicateDetails, fmt.Sprintf("Row %d (MSISDN %s): Exact duplicate event already exists in database (ID: %s).", csvRowCount, msisdn, existingEvent.ID.String()))
			continue
		} else if result.Error != gorm.ErrRecordNotFound {
			log.Printf("DEBUG: Row %d (MSISDN %s): DB error checking for duplicate event: %v. Skipping row.\n", csvRowCount, msisdn, result.Error)
			errorMessages = append(errorMessages, fmt.Sprintf("Row %d (MSISDN %s): DB error checking for duplicate event: %s", csvRowCount, msisdn, result.Error.Error()))
			continue
		}
		log.Printf("DEBUG: Row %d (MSISDN %s): Event is unique. Adding to batch for creation.\n", csvRowCount, msisdn)

		participantEventsToCreate = append(participantEventsToCreate, models.ParticipantEvent{
			MSISDN:               msisdn,
			Amount:               amountStr,
			OptInStatus:          optInStatusStr,
			PointsEarned:         pointsEarned,
			TransactionTimestamp: transactionTime,
			UploadAuditID:        auditEntry.ID,
		})

		if pMaster, exists := participantMasterMap[msisdn]; !exists {
			participantMasterMap[msisdn] = models.Participant{MSISDN: msisdn, OptInDate: transactionTime}
		} else {
		    if pMaster.OptInDate == nil || (transactionTime != nil && transactionTime.Before(*pMaster.OptInDate)) {
		        pMaster.OptInDate = transactionTime
                participantMasterMap[msisdn] = pMaster
		    }
		}
	}
	log.Println("DEBUG: Finished CSV row processing loop")

	for _, p := range participantMasterMap {
	    participantsToUpsert = append(participantsToUpsert, p)
	}

	auditEntry.RecordCount = csvRowCount - 1
	log.Printf("DEBUG: Total data rows in CSV (excluding header): %d\n", auditEntry.RecordCount)
	log.Printf("DEBUG: Number of unique events to create: %d\n", len(participantEventsToCreate))
	log.Printf("DEBUG: Number of participant master records to upsert: %d\n", len(participantsToUpsert))

	log.Println("DEBUG: Starting database transaction")
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if len(participantsToUpsert) > 0 {
			log.Printf("DEBUG: Upserting %d participant master records\n", len(participantsToUpsert))
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "msisdn"}}, DoUpdates: clause.AssignmentColumns([]string{"opt_in_date", "updated_at"})}).Create(&participantsToUpsert).Error; err != nil {
				log.Printf("DEBUG: Error upserting participant master records: %v\n", err)
				return fmt.Errorf("failed to upsert participant master records: %w", err)
			}
			log.Println("DEBUG: Participant master records upserted successfully")
		}

		if len(participantEventsToCreate) > 0 {
			log.Printf("DEBUG: Creating %d participant events in batches\n", len(participantEventsToCreate))
			if err := tx.CreateInBatches(&participantEventsToCreate, 100).Error; err != nil {
				log.Printf("DEBUG: Error creating participant events: %v\n", err)
				return fmt.Errorf("failed to create participant events: %w", err)
			}
			successfullyImportedCount = len(participantEventsToCreate)
			log.Printf("DEBUG: Participant events created successfully. Count: %d\n", successfullyImportedCount)
		}
		return nil
	})

	if err != nil {
		log.Printf("DEBUG: Database transaction error: %v\n", err)
		errorMessages = append(errorMessages, "Database transaction error: "+err.Error())
		auditEntry.Status = "Failed"
	} else {
		log.Println("DEBUG: Database transaction successful")
		if len(errorMessages) > 0 {
			auditEntry.Status = "Partial Success"
		} else if successfullyImportedCount == 0 && auditEntry.RecordCount > 0 && duplicatesSkippedCount == auditEntry.RecordCount {
            auditEntry.Status = "Success" // All records were duplicates
        } else if successfullyImportedCount == 0 && auditEntry.RecordCount == 0 && duplicatesSkippedCount == 0 && len(errorMessages) == 0 {
            auditEntry.Status = "Failed" // No data rows found or processed, and no specific errors (e.g. empty file after header)
            errorMessages = append(errorMessages, "No data rows found or processed from CSV.")
        } else {
			auditEntry.Status = "Success"
		}
	}
	log.Printf("DEBUG: Final Audit Status: %s, Imported: %d, Duplicates Skipped: %d, Errors: %d\n", auditEntry.Status, successfullyImportedCount, duplicatesSkippedCount, len(errorMessages))

	auditEntry.SuccessfullyImported = successfullyImportedCount
	auditEntry.DuplicatesSkipped = duplicatesSkippedCount
	notes := strings.Join(errorMessages, "; ")
	if len(skippedDuplicateDetails) > 0 {
		notes += "; Skipped Duplicates: " + strings.Join(skippedDuplicateDetails, ", ")
	}
    auditEntry.Notes = notes

	if errSaveAudit := config.DB.Save(&auditEntry).Error; errSaveAudit != nil {
		log.Printf("CRITICAL: Failed to update final audit entry %s: %v\n", auditEntry.ID, errSaveAudit)
	}

	responseStatus := http.StatusOK
	if auditEntry.Status == "Failed" {
		responseStatus = http.StatusBadRequest
	} else if auditEntry.Status == "Partial Success" {
	    responseStatus = http.StatusMultiStatus
	}

	log.Printf("DEBUG: Responding to client with status %d. Audit ID: %s\n", responseStatus, auditEntry.ID.String())
	c.JSON(responseStatus, gin.H{
		"message":                         "File processing complete.",
		"audit_id":                        auditEntry.ID,
		"status":                          auditEntry.Status,
		"total_data_rows_processed":       auditEntry.RecordCount,
		"successfully_imported_rows":      auditEntry.SuccessfullyImported,
		"duplicates_skipped_count":        auditEntry.DuplicatesSkipped,
		"processing_error_messages":       errorMessages,
        "skipped_duplicate_event_details": skippedDuplicateDetails,
	})
	log.Println("DEBUG: HandleParticipantUpload finished")
}

