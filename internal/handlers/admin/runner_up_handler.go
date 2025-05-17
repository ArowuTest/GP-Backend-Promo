package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InvokeRunnerUp godoc
// @Summary Invoke a runner-up for a prize
// @Description Promotes a runner-up to winner status when a previous winner forfeits
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body InvokeRunnerUpRequest true "Invoke runner-up request"
// @Success 200 {object} InvokeRunnerUpResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws/invoke-runner-up [post]
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	var req InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate request
	if req.WinnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "winner_id is required"})
		return
	}

	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid winner_id format"})
		return
	}

	// Get the current winner
	var winner models.DrawWinner
	if err := config.DB.First(&winner, "id = ?", winnerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Winner not found"})
		return
	}

	// Check if winner is already forfeited
	if winner.Status == "Forfeited" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Winner has already forfeited"})
		return
	}

	// Get the draw
	var draw models.Draw
	if err := config.DB.First(&draw, "id = ?", winner.DrawID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		return
	}

	// Begin transaction
	tx := config.DB.Begin()

	// Update current winner status to forfeited
	winner.Status = "Forfeited"
	winner.Notes = req.Reason
	if err := tx.Save(&winner).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update winner status: " + err.Error()})
		return
	}

	// Find the next eligible runner-up for this prize tier
	var runnerUp models.DrawWinner
	if err := tx.Where("draw_id = ? AND prize_tier_id = ? AND is_runner_up = ? AND status = ?",
		winner.DrawID, winner.PrizeTierID, true, "PendingInvocation").
		Order("runner_up_rank ASC").
		First(&runnerUp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "No eligible runner-up found for this prize"})
		return
	}

	// Update runner-up to winner status
	runnerUp.IsRunnerUp = false
	runnerUp.Status = "PendingNotification"
	if err := tx.Save(&runnerUp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner-up status: " + err.Error()})
		return
	}

	// Create audit log for this action
	userID, _ := c.Get("userID")
	auditDetails := map[string]interface{}{
		"winner_id":      winner.ID.String(),
		"runner_up_id":   runnerUp.ID.String(),
		"draw_id":        draw.ID.String(),
		"prize_tier_id":  winner.PrizeTierID,
		"reason":         req.Reason,
		"forfeited_msisdn": winner.MSISDN,
		"promoted_msisdn":  runnerUp.MSISDN,
	}
	
	detailsJSON, _ := json.Marshal(auditDetails)
	
	auditLog := models.SystemAuditLog{
		UserID:        userID.(uuid.UUID),
		ActionType:    "INVOKE_RUNNER_UP",
		ResourceType:  "DRAW_WINNER",
		ResourceID:    winner.ID.String(),
		Description:   fmt.Sprintf("Invoked runner-up for forfeited winner in draw %s", draw.ID.String()),
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		ActionDetails: string(detailsJSON),
	}
	
	if err := tx.Create(&auditLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit log: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// Try to send notification to the new winner
	go func() {
		// This would be implemented to use the SMS gateway
		// For now, just log the attempt
		fmt.Printf("Notification would be sent to new winner %s for prize %s\n", 
			runnerUp.MSISDN, runnerUp.PrizeTierID)
	}()

	// Return success response
	c.JSON(http.StatusOK, InvokeRunnerUpResponse{
		Message: "Runner-up successfully invoked",
		ForfeitedWinner: WinnerResponse{
			ID:     winner.ID.String(),
			MSISDN: winner.MSISDN,
			Status: winner.Status,
		},
		PromotedRunnerUp: WinnerResponse{
			ID:     runnerUp.ID.String(),
			MSISDN: runnerUp.MSISDN,
			Status: runnerUp.Status,
		},
	})
}

// InvokeRunnerUpRequest defines the request structure for invoking a runner-up
type InvokeRunnerUpRequest struct {
	WinnerID string `json:"winner_id" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

// InvokeRunnerUpResponse defines the response structure for invoking a runner-up
type InvokeRunnerUpResponse struct {
	Message          string         `json:"message"`
	ForfeitedWinner  WinnerResponse `json:"forfeited_winner"`
	PromotedRunnerUp WinnerResponse `json:"promoted_runner_up"`
}

// WinnerResponse defines the winner data in responses
type WinnerResponse struct {
	ID     string `json:"id"`
	MSISDN string `json:"msisdn"`
	Status string `json:"status"`
}
