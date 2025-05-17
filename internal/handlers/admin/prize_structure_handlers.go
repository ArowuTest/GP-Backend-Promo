package admin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause" // Correct import for GORM clauses
)

// AdminHandlers handles all admin-related operations
type AdminHandlers struct {
	db *gorm.DB
	auditService interface{} // This would be your AuditService interface
	drawDataService interface{} // This would be your DrawDataService interface
}

// NewAdminHandlers creates a new AdminHandlers instance
func NewAdminHandlers(
	auditService interface{},
	drawDataService interface{},
	db *gorm.DB,
) *AdminHandlers {
	return &AdminHandlers{
		db: db,
		auditService: auditService,
		drawDataService: drawDataService,
	}
}

// CreatePrizeStructureRequest defines the structure for creating a prize structure
type CreatePrizeStructureRequest struct {
	Name           string                      `json:"name" binding:"required"`
	Description    string                      `json:"description,omitempty"`
	IsActive       bool                        `json:"is_active"` // Default is true in model
	ValidFrom      *time.Time                  `json:"valid_from,omitempty"`
	ValidTo        *time.Time                  `json:"valid_to,omitempty"`
	Prizes         []models.CreatePrizeRequest `json:"prizes" binding:"required,dive"`
	ApplicableDays []string                    `json:"applicable_days,omitempty"` // Added field for applicable days
}

// CreatePrizeStructure handles the creation of a new prize structure
func (h *AdminHandlers) CreatePrizeStructure(c *gin.Context) {
	// Enhanced logging for debugging
	fmt.Println("CreatePrizeStructure handler called")
	
	// Log request headers for debugging
	fmt.Println("Request Headers:")
	for k, v := range c.Request.Header {
		fmt.Printf("%s: %v\n", k, v)
	}
	
	// Read raw request body for debugging
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body: " + err.Error()})
		return
	}
	fmt.Printf("Raw request body: %s\n", string(bodyBytes))
	
	// Reset request body for binding
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	var req CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Log the received payload for debugging
	fmt.Printf("Received CreatePrizeStructure payload: %+v\n", req)
	fmt.Printf("Number of prizes in payload: %d\n", len(req.Prizes))
	for i, prize := range req.Prizes {
		fmt.Printf("Prize %d: %+v\n", i, prize)
	}
	
	// Log applicable days
	fmt.Printf("Applicable days: %v\n", req.ApplicableDays)

	// Validate dates
	if req.ValidFrom != nil && req.ValidTo != nil && req.ValidTo.Before(*req.ValidFrom) {
		fmt.Println("Date validation error: ValidTo before ValidFrom")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidTo cannot be before ValidFrom"})
		return
	}

	// Ensure ValidFrom is not zero if provided
	if req.ValidFrom != nil && req.ValidFrom.IsZero() {
		fmt.Println("Date validation error: ValidFrom is zero")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ValidFrom date"})
		return
	}

	// Ensure ValidTo is not zero if provided
	if req.ValidTo != nil && req.ValidTo.IsZero() {
		fmt.Println("Date validation error: ValidTo is zero")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ValidTo date"})
		return
	}

	adminIDClaim, exists := c.Get("userID")
	if !exists {
		fmt.Println("Auth error: Admin user ID not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}

	adminIDStr, ok := adminIDClaim.(string)
	if !ok {
		fmt.Println("Auth error: Admin user ID in token is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin user ID in token is not a string"})
		return
	}

	parsedAdminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		fmt.Printf("Auth error: Invalid admin user ID format: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	// Set default applicable days if not provided
	if len(req.ApplicableDays) == 0 {
		fmt.Println("Setting default applicable days (all days)")
		req.ApplicableDays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	}

	// Derive day_type from applicable_days
	dayType := deriveDayTypeFromApplicableDays(req.ApplicableDays)
	fmt.Printf("Derived day_type: %s from applicable_days: %v\n", dayType, req.ApplicableDays)

	var createdPrizeStructure models.PrizeStructure
	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		prizeStructure := models.PrizeStructure{
			Name:             req.Name,
			Description:      req.Description,
			IsActive:         req.IsActive, // Model has default true, this will override if false
			ValidFrom:        req.ValidFrom,
			ValidTo:          req.ValidTo,
			CreatedByAdminID: parsedAdminID,
			DayType:          dayType, // Set the derived day_type
			ApplicableDays:   req.ApplicableDays, // Store applicable_days in virtual field for response
		}

		if err := tx.Create(&prizeStructure).Error; err != nil {
			fmt.Printf("Error creating prize structure: %v\n", err)
			return fmt.Errorf("failed to create prize structure: %w", err)
		}

		fmt.Printf("Created prize structure with ID: %s\n", prizeStructure.ID)

		// Create prizes one by one and check for errors
		for i, prizeReq := range req.Prizes {
			prize := models.Prize{
				PrizeStructureID:   prizeStructure.ID,
				Name:               prizeReq.Name,
				Value:              prizeReq.Value,
				PrizeType:          prizeReq.PrizeType,
				Quantity:           prizeReq.Quantity,
				Order:              prizeReq.Order,
				NumberOfRunnerUps:  prizeReq.NumberOfRunnerUps,
			}

			if err := tx.Create(&prize).Error; err != nil {
				fmt.Printf("Error creating prize tier %d: %v\n", i, err)
				return fmt.Errorf("failed to create prize tier %d (%s): %w", i, prizeReq.Name, err)
			}
			fmt.Printf("Created prize tier %d with ID: %s\n", i, prize.ID)
		}

		// Reload the prize structure with prizes to return in response
		if err := tx.Preload("Prizes").First(&prizeStructure, prizeStructure.ID).Error; err != nil {
			fmt.Printf("Error reloading prize structure: %v\n", err)
			return fmt.Errorf("failed to reload prize structure with prizes: %w", err)
		}

		createdPrizeStructure = prizeStructure
		return nil
	})

	if txErr != nil {
		fmt.Printf("Transaction error in CreatePrizeStructure: %v\n", txErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	// Format dates for response
	formatDatesForResponse(&createdPrizeStructure)

	// Ensure applicable_days is populated in the response
	createdPrizeStructure.ApplicableDays = req.ApplicableDays

	fmt.Println("Prize structure created successfully")
	c.JSON(http.StatusCreated, createdPrizeStructure)
}

// ListPrizeStructures handles listing all prize structures
func (h *AdminHandlers) ListPrizeStructures(c *gin.Context) {
	fmt.Println("ListPrizeStructures handler called")
	
	var prizeStructures []models.PrizeStructure
	result := h.db.Preload("Prizes").Order("created_at desc").Find(&prizeStructures)
	if result.Error != nil {
		fmt.Printf("Error retrieving prize structures: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structures: " + result.Error.Error()})
		return
	}

	fmt.Printf("Found %d prize structures\n", len(prizeStructures))

	// Process each prize structure to populate applicable_days from day_type
	for i := range prizeStructures {
		prizeStructures[i].ApplicableDays = getApplicableDaysFromDayType(prizeStructures[i].DayType)
		formatDatesForResponse(&prizeStructures[i])
	}

	c.JSON(http.StatusOK, prizeStructures)
}

// GetPrizeStructure handles retrieving a single prize structure by ID
func (h *AdminHandlers) GetPrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	fmt.Printf("GetPrizeStructure handler called for ID: %s\n", structureIDStr)
	
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		fmt.Printf("Invalid prize structure ID format: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var prizeStructure models.PrizeStructure
	result := h.db.Preload("Prizes").First(&prizeStructure, "id = ?", structureID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Printf("Prize structure not found for ID: %s\n", structureIDStr)
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
			return
		}
		fmt.Printf("Error retrieving prize structure: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + result.Error.Error()})
		return
	}

	// Populate applicable_days from day_type
	prizeStructure.ApplicableDays = getApplicableDaysFromDayType(prizeStructure.DayType)
	
	// Format dates for response
	formatDatesForResponse(&prizeStructure)

	fmt.Printf("Retrieved prize structure with ID: %s\n", structureIDStr)
	c.JSON(http.StatusOK, prizeStructure)
}

// UpdatePrizeStructureRequest defines the structure for updating a prize structure
type UpdatePrizeStructureRequest struct {
	Name           *string                      `json:"name,omitempty"`
	Description    *string                      `json:"description,omitempty"`
	IsActive       *bool                        `json:"is_active,omitempty"`
	ValidFrom      *time.Time                   `json:"valid_from,omitempty"`
	ValidTo        **time.Time                  `json:"valid_to,omitempty"` // Pointer to pointer for explicit null
	Prizes         *[]models.CreatePrizeRequest `json:"prizes,omitempty"`
	ApplicableDays *[]string                    `json:"applicable_days,omitempty"` // Added field for applicable days
}

// UpdatePrizeStructure handles updating an existing prize structure
func (h *AdminHandlers) UpdatePrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	fmt.Printf("UpdatePrizeStructure handler called for ID: %s\n", structureIDStr)
	
	// Log request headers for debugging
	fmt.Println("Request Headers:")
	for k, v := range c.Request.Header {
		fmt.Printf("%s: %v\n", k, v)
	}
	
	// Read raw request body for debugging
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body: " + err.Error()})
		return
	}
	fmt.Printf("Raw request body: %s\n", string(bodyBytes))
	
	// Reset request body for binding
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		fmt.Printf("Invalid prize structure ID format: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var req UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Log the received payload for debugging
	fmt.Printf("Received UpdatePrizeStructure payload for ID %s: %+v\n", structureIDStr, req)
	if req.Prizes != nil {
		fmt.Printf("Number of prizes in update payload: %d\n", len(*req.Prizes))
		for i, prize := range *req.Prizes {
			fmt.Printf("Prize %d: %+v\n", i, prize)
		}
	}
	
	// Log applicable days if provided
	if req.ApplicableDays != nil {
		fmt.Printf("Applicable days: %v\n", *req.ApplicableDays)
	}

	// Validate ValidFrom if provided
	if req.ValidFrom != nil && req.ValidFrom.IsZero() {
		fmt.Println("Date validation error: ValidFrom is zero")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ValidFrom date"})
		return
	}

	// Validate ValidTo if provided and not null
	if req.ValidTo != nil && *req.ValidTo != nil && (*req.ValidTo).IsZero() {
		fmt.Println("Date validation error: ValidTo is zero")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ValidTo date"})
		return
	}

	var updatedPrizeStructure models.PrizeStructure
	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		var prizeStructure models.PrizeStructure
		if err := tx.Preload("Prizes").First(&prizeStructure, "id = ?", structureID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Printf("Prize structure not found for ID: %s\n", structureIDStr)
				return errors.New("prize structure not found")
			}
			fmt.Printf("Error finding prize structure: %v\n", err)
			return fmt.Errorf("failed to find prize structure: %w", err)
		}

		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		if req.ValidFrom != nil {
			updates["valid_from"] = *req.ValidFrom
		}
		if req.ValidTo != nil {
			if *req.ValidTo == nil {
				updates["valid_to"] = clause.Expr{SQL: "NULL"} // Correct way to set NULL with GORM
			} else {
				updates["valid_to"] = **req.ValidTo
			}
		}
		if req.ApplicableDays != nil {
			// Derive day_type from applicable_days
			dayType := deriveDayTypeFromApplicableDays(*req.ApplicableDays)
			updates["day_type"] = dayType
			prizeStructure.ApplicableDays = *req.ApplicableDays // Update virtual field
			fmt.Printf("Updating day_type to: %s from applicable_days: %v\n", dayType, *req.ApplicableDays)
		}

		currentValidFrom := prizeStructure.ValidFrom
		if val, ok := updates["valid_from"].(time.Time); ok {
			currentValidFrom = &val
		}

		var currentValidTo *time.Time = prizeStructure.ValidTo
		if val, ok := updates["valid_to"].(time.Time); ok {
			currentValidTo = &val
		} else if _, ok := updates["valid_to"].(clause.Expr); ok {
			currentValidTo = nil // Being set to NULL
		}

		if currentValidFrom != nil && currentValidTo != nil && currentValidTo.Before(*currentValidFrom) {
			fmt.Println("Date validation error: ValidTo before ValidFrom")
			return errors.New("ValidTo cannot be before ValidFrom")
		}

		if len(updates) > 0 {
			if err := tx.Model(&prizeStructure).Updates(updates).Error; err != nil {
				fmt.Printf("Error updating prize structure fields: %v\n", err)
				return fmt.Errorf("failed to update prize structure details: %w", err)
			}
			fmt.Printf("Updated prize structure fields: %v\n", updates)
		}

		if req.Prizes != nil {
			// First delete existing prizes
			if err := tx.Where("prize_structure_id = ?", prizeStructure.ID).Delete(&models.Prize{}).Error; err != nil {
				fmt.Printf("Error deleting existing prizes: %v\n", err)
				return fmt.Errorf("failed to delete existing prizes: %w", err)
			}
			fmt.Printf("Deleted existing prizes for structure ID: %s\n", prizeStructure.ID)

			// Then create new prizes
			for i, prizeReq := range *req.Prizes {
				newPrize := models.Prize{
					PrizeStructureID:  prizeStructure.ID,
					Name:              prizeReq.Name,
					Value:             prizeReq.Value,
					PrizeType:         prizeReq.PrizeType,
					Quantity:          prizeReq.Quantity,
					Order:             prizeReq.Order,
					NumberOfRunnerUps: prizeReq.NumberOfRunnerUps,
				}

				if err := tx.Create(&newPrize).Error; err != nil {
					fmt.Printf("Error creating prize tier %d: %v\n", i, err)
					return fmt.Errorf("failed to create prize tier %d (%s): %w", i, prizeReq.Name, err)
				}
				fmt.Printf("Created new prize tier %d with ID: %s\n", i, newPrize.ID)
			}
		}

		// Reload the prize structure with prizes to return in response
		if err := tx.Preload("Prizes").First(&prizeStructure, prizeStructure.ID).Error; err != nil {
			fmt.Printf("Error reloading prize structure: %v\n", err)
			return fmt.Errorf("failed to reload prize structure with prizes: %w", err)
		}

		updatedPrizeStructure = prizeStructure
		return nil
	})

	if txErr != nil {
		fmt.Printf("Transaction error in UpdatePrizeStructure: %v\n", txErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	// Format dates for response
	formatDatesForResponse(&updatedPrizeStructure)

	// Ensure applicable_days is populated in the response
	if req.ApplicableDays != nil {
		updatedPrizeStructure.ApplicableDays = *req.ApplicableDays
	} else {
		updatedPrizeStructure.ApplicableDays = getApplicableDaysFromDayType(updatedPrizeStructure.DayType)
	}

	fmt.Printf("Prize structure updated successfully: %+v\n", updatedPrizeStructure)
	c.JSON(http.StatusOK, updatedPrizeStructure)
}

// DeletePrizeStructure handles deleting a prize structure by ID
func (h *AdminHandlers) DeletePrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	fmt.Printf("DeletePrizeStructure handler called for ID: %s\n", structureIDStr)
	
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		fmt.Printf("Invalid prize structure ID format: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var prizeStructure models.PrizeStructure
	if err := h.db.First(&prizeStructure, "id = ?", structureID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Prize structure not found for ID: %s\n", structureIDStr)
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
			return
		}
		fmt.Printf("Error finding prize structure: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find prize structure: " + err.Error()})
		return
	}

	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated prizes first
		if err := tx.Where("prize_structure_id = ?", structureID).Delete(&models.Prize{}).Error; err != nil {
			fmt.Printf("Error deleting associated prizes: %v\n", err)
			return fmt.Errorf("failed to delete associated prizes: %w", err)
		}
		fmt.Printf("Deleted prizes for structure ID: %s\n", structureID)

		// Then delete the prize structure
		if err := tx.Delete(&prizeStructure).Error; err != nil {
			fmt.Printf("Error deleting prize structure: %v\n", err)
			return fmt.Errorf("failed to delete prize structure: %w", err)
		}
		fmt.Printf("Deleted prize structure with ID: %s\n", structureID)

		return nil
	})

	if txErr != nil {
		fmt.Printf("Transaction error in DeletePrizeStructure: %v\n", txErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	fmt.Println("Prize structure deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Prize structure deleted successfully"})
}

// Helper functions

// deriveDayTypeFromApplicableDays converts a slice of day names to a day_type string
func deriveDayTypeFromApplicableDays(days []string) string {
	if len(days) == 0 {
		return "all" // Default to all days if none specified
	}

	// Check if all days are included
	if len(days) == 7 {
		return "all"
	}

	// Check if only weekdays
	weekdays := map[string]bool{"Mon": true, "Tue": true, "Wed": true, "Thu": true, "Fri": true}
	isWeekdays := true
	for _, day := range days {
		if !weekdays[day] {
			isWeekdays = false
			break
		}
	}
	if isWeekdays && len(days) == 5 {
		return "weekday"
	}

	// Check if only weekends
	weekends := map[string]bool{"Sat": true, "Sun": true}
	isWeekends := true
	for _, day := range days {
		if !weekends[day] {
			isWeekends = false
			break
		}
	}
	if isWeekends && len(days) == 2 {
		return "weekend"
	}

	// Otherwise, it's a custom set of days
	return "custom"
}

// getApplicableDaysFromDayType converts a day_type string to a slice of day names
func getApplicableDaysFromDayType(dayType string) []string {
	switch dayType {
	case "all":
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	case "weekday":
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	case "weekend":
		return []string{"Sat", "Sun"}
	case "custom":
		// For custom, we would need to retrieve the actual days from somewhere
		// For now, return an empty slice
		return []string{}
	default:
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} // Default to all days
	}
}

// formatDatesForResponse ensures dates are in a consistent format for API responses
func formatDatesForResponse(prizeStructure *models.PrizeStructure) {
	// This function ensures dates are formatted consistently
	// It's a no-op in Go since the JSON marshaler handles time.Time correctly,
	// but it's here for symmetry with the frontend code
}

// Additional handler methods for other admin operations would go here
// For example: ExecuteDraw, GetDrawEligibilityStats, ListDraws, GetDraw, etc.
