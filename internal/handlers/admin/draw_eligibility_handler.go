package admin

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
)

// GetEligibilityStats godoc
// @Summary Get eligibility statistics for a draw
// @Description Retrieves the number of eligible participants and total entries for a draw on a specific date
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param drawDate query string true "Draw date in YYYY-MM-DD format"
// @Param prize_structure_id query string false "Prize structure ID (UUID)"
// @Success 200 {object} services.DrawEligibilityStats
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws/eligibility-stats [get]
func (h *DrawHandler) GetEligibilityStats(c *gin.Context) {
	drawDateStr := c.Query("drawDate")
	if drawDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "drawDate query parameter is required"})
		return
	}

	drawDate, err := time.Parse("2006-01-02", drawDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawDate format. Expected YYYY-MM-DD"})
		return
	}

	prizeStructureID := c.Query("prize_structure_id")
	if prizeStructureID == "" {
		// If no prize structure ID is provided, we'll use a default or the first active one
		// For now, just return an error
		c.JSON(http.StatusBadRequest, gin.H{"error": "prize_structure_id query parameter is required"})
		return
	}

	participants, err := h.DrawDataService.GetEligibleParticipants(drawDate, prizeStructureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get eligible participants: " + err.Error()})
		return
	}

	// Calculate total points
	totalPoints := 0
	for _, p := range participants {
		totalPoints += p.TotalPoints
	}

	c.JSON(http.StatusOK, gin.H{
		"totalEligibleMSISDNs": len(participants),
		"totalEntries":         totalPoints,
	})
}

// DrawEligibilityStats defines the structure for the eligibility stats API response
type DrawEligibilityStats struct {
	TotalEligibleMSISDNs int `json:"totalEligibleMSISDNs"`
	TotalEntries         int `json:"totalEntries"`
}

// DrawHandler handles draw related requests
type DrawHandler struct {
	DrawDataService services.DrawDataService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(dds services.DrawDataService) *DrawHandler {
	return &DrawHandler{DrawDataService: dds}
}
