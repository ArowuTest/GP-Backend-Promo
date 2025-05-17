package admin

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DrawEligibilityHandler handles operations related to draw eligibility
type DrawEligibilityHandler struct {
	db             *gorm.DB
	drawDataService services.DrawDataService
}

// NewDrawEligibilityHandler creates a new DrawEligibilityHandler
func NewDrawEligibilityHandler(db *gorm.DB, drawDataService services.DrawDataService) *DrawEligibilityHandler {
	return &DrawEligibilityHandler{
		db:             db,
		drawDataService: drawDataService,
	}
}

// GetDrawEligibilityStats returns statistics about eligible participants for a draw
func (h *DrawEligibilityHandler) GetDrawEligibilityStats(c *gin.Context) {
	// Parse date parameter
	dateStr := c.Query("date")
	var targetDate time.Time
	var err error

	if dateStr == "" {
		// Default to today if no date provided
		targetDate = time.Now()
	} else {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	// Get day of week for the target date
	dayOfWeek := targetDate.Weekday().String()[:3] // Mon, Tue, etc.

	// Get eligible prize structures for this day
	var eligiblePrizeStructures []models.PrizeStructure
	dayTypeConditions := []string{"all"}

	// Add specific day type based on weekday/weekend
	if dayOfWeek == "Sat" || dayOfWeek == "Sun" {
		dayTypeConditions = append(dayTypeConditions, "weekend")
	} else {
		dayTypeConditions = append(dayTypeConditions, "weekday")
	}

	// Add custom day type that might include this specific day
	dayTypeConditions = append(dayTypeConditions, "custom")

	// Query for eligible prize structures
	if err := h.db.Where("is_active = ? AND (valid_from IS NULL OR valid_from <= ?) AND (valid_to IS NULL OR valid_to >= ?) AND day_type IN ?",
		true, targetDate, targetDate, dayTypeConditions).
		Preload("Prizes").
		Find(&eligiblePrizeStructures).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve eligible prize structures: " + err.Error()})
		return
	}

	// Filter custom day types to only include those that match this specific day
	var filteredPrizeStructures []models.PrizeStructure
	for _, ps := range eligiblePrizeStructures {
		if ps.DayType != "custom" || contains(getApplicableDaysFromDayType(ps.DayType), dayOfWeek) {
			filteredPrizeStructures = append(filteredPrizeStructures, ps)
		}
	}

	// Get eligible participant count from draw data service
	eligibleCount, err := h.drawDataService.GetEligibleParticipantCount(targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve eligible participant count: " + err.Error()})
		return
	}

	// Construct response
	response := gin.H{
		"date":                   targetDate.Format("2006-01-02"),
		"day_of_week":            dayOfWeek,
		"eligible_participants":  eligibleCount,
		"eligible_prize_structures": filteredPrizeStructures,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getApplicableDaysFromDayType converts a day_type string to a slice of day names
// This is a duplicate of the function in prize_structure_handlers.go and should be moved to a common utility package
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
