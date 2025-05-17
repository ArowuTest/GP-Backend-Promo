package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DrawHandler handles operations related to draws
type DrawHandler struct {
	db             *gorm.DB
	drawDataService services.DrawDataService
	auditService    *services.AuditService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(db *gorm.DB, drawDataService services.DrawDataService, auditService *services.AuditService) *DrawHandler {
	return &DrawHandler{
		db:             db,
		drawDataService: drawDataService,
		auditService:    auditService,
	}
}

// InvokeRunnerUpRequest defines the structure for invoking a runner-up
type InvokeRunnerUpRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ExecuteDrawRequest defines the structure for executing a draw
type ExecuteDrawRequest struct {
	Date          string `json:"date" binding:"required"`
	PrizeStructureID string `json:"prize_structure_id" binding:"required"`
}

// ExecuteDraw handles the execution of a draw
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Parse date
	drawDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	// Get admin ID from context
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	// Verify prize structure exists and is eligible for the draw date
	var prizeStructure models.PrizeStructure
	if err := h.db.Preload("Prizes").First(&prizeStructure, prizeStructureID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + err.Error()})
		}
		return
	}

	// Check if prize structure is active
	if !prizeStructure.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prize structure is not active"})
		return
	}

	// Check if prize structure is valid for the draw date
	if prizeStructure.ValidFrom != nil && drawDate.Before(*prizeStructure.ValidFrom) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Draw date is before prize structure valid from date"})
		return
	}
	if prizeStructure.ValidTo != nil && drawDate.After(*prizeStructure.ValidTo) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Draw date is after prize structure valid to date"})
		return
	}

	// Check if prize structure is applicable for the day of the week
	dayOfWeek := drawDate.Weekday().String()[:3] // Mon, Tue, etc.
	applicableDays := getApplicableDaysFromDayType(prizeStructure.DayType)
	if !contains(applicableDays, dayOfWeek) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prize structure is not applicable for this day of the week"})
		return
	}

	// Get eligible participants for the draw
	eligibleParticipants, err := h.drawDataService.GetEligibleParticipants(drawDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve eligible participants: " + err.Error()})
		return
	}

	if len(eligibleParticipants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No eligible participants for this draw date"})
		return
	}

	// Create a new draw record
	draw := models.Draw{
		Date:             drawDate,
		PrizeStructureID: prizeStructureID,
		ExecutedByAdminID: adminID,
		Status:           "completed",
		ParticipantCount: len(eligibleParticipants),
	}

	// Begin transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + tx.Error.Error()})
		return
	}

	// Create draw record
	if err := tx.Create(&draw).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create draw record: " + err.Error()})
		return
	}

	// Execute draw for each prize
	var winners []models.Winner
	var runnerUps []models.RunnerUp
	usedMSISDNs := make(map[string]bool) // Track MSISDNs that have already won

	// Sort prizes by order
	sortedPrizes := prizeStructure.Prizes
	// Implement sorting logic here if needed

	for _, prize := range sortedPrizes {
		// Skip if prize quantity is 0
		if prize.Quantity <= 0 {
			continue
		}

		// Select winners for this prize
		prizeWinners, prizeRunnerUps, err := h.selectWinnersAndRunnerUps(eligibleParticipants, prize.Quantity, prize.NumberOfRunnerUps, usedMSISDNs)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select winners: " + err.Error()})
			return
		}

		// Create winner records
		for i, participant := range prizeWinners {
			winner := models.Winner{
				DrawID:       draw.ID,
				PrizeID:      prize.ID,
				MSISDN:       participant.MSISDN,
				Points:       participant.Points,
				Status:       "pending",
				WinnerNumber: i + 1,
			}
			if err := tx.Create(&winner).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create winner record: " + err.Error()})
				return
			}
			winners = append(winners, winner)
			usedMSISDNs[participant.MSISDN] = true
		}

		// Create runner-up records
		for i, participant := range prizeRunnerUps {
			runnerUp := models.RunnerUp{
				DrawID:         draw.ID,
				PrizeID:        prize.ID,
				MSISDN:         participant.MSISDN,
				Points:         participant.Points,
				Status:         "pending",
				RunnerUpNumber: i + 1,
			}
			if err := tx.Create(&runnerUp).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create runner-up record: " + err.Error()})
				return
			}
			runnerUps = append(runnerUps, runnerUp)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// Log audit event
	auditEvent := models.AuditLog{
		AdminID:     adminID,
		Action:      "execute_draw",
		EntityType:  "draw",
		EntityID:    draw.ID.String(),
		Description: fmt.Sprintf("Executed draw for date %s with prize structure %s", req.Date, prizeStructure.Name),
	}
	h.auditService.LogAuditEvent(auditEvent)

	// Return response with draw details
	response := gin.H{
		"draw_id":           draw.ID,
		"date":              draw.Date.Format("2006-01-02"),
		"prize_structure":   prizeStructure,
		"participant_count": draw.ParticipantCount,
		"winners":           winners,
		"runner_ups":        runnerUps,
	}

	c.JSON(http.StatusOK, response)
}

// ListDraws handles listing all draws
func (h *DrawHandler) ListDraws(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get total count
	var totalCount int64
	if err := h.db.Model(&models.Draw{}).Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count draws: " + err.Error()})
		return
	}

	// Get draws with pagination
	var draws []models.Draw
	if err := h.db.Preload("PrizeStructure").
		Preload("ExecutedByAdmin").
		Order("date desc").
		Offset(offset).
		Limit(limit).
		Find(&draws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draws: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"draws":       draws,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	})
}

// GetDraw handles retrieving a single draw by ID
func (h *DrawHandler) GetDraw(c *gin.Context) {
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw ID format"})
		return
	}

	var draw models.Draw
	if err := h.db.Preload("PrizeStructure").
		Preload("PrizeStructure.Prizes").
		Preload("ExecutedByAdmin").
		First(&draw, drawID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw: " + err.Error()})
		}
		return
	}

	// Get winners for this draw
	var winners []models.Winner
	if err := h.db.Where("draw_id = ?", drawID).
		Preload("Prize").
		Find(&winners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winners: " + err.Error()})
		return
	}

	// Get runner-ups for this draw
	var runnerUps []models.RunnerUp
	if err := h.db.Where("draw_id = ?", drawID).
		Preload("Prize").
		Find(&runnerUps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner-ups: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"draw":       draw,
		"winners":    winners,
		"runner_ups": runnerUps,
	})
}

// InvokeRunnerUp handles invoking a runner-up when a winner cannot claim their prize
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	winnerIDStr := c.Param("id")
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid winner ID format"})
		return
	}

	var req InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Get admin ID from context
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + tx.Error.Error()})
		return
	}

	// Get winner record
	var winner models.Winner
	if err := tx.First(&winner, winnerID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Winner not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winner: " + err.Error()})
		}
		return
	}

	// Check if winner status allows invoking runner-up
	if winner.Status != "pending" && winner.Status != "notified" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot invoke runner-up for winner with status: " + winner.Status})
		return
	}

	// Get next available runner-up
	var runnerUp models.RunnerUp
	if err := tx.Where("draw_id = ? AND prize_id = ? AND status = ?", winner.DrawID, winner.PrizeID, "pending").
		Order("runner_up_number asc").
		First(&runnerUp).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No available runner-ups for this prize"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner-up: " + err.Error()})
		}
		return
	}

	// Update winner status to forfeited
	if err := tx.Model(&winner).Updates(map[string]interface{}{
		"status":           "forfeited",
		"forfeit_reason":   req.Reason,
		"forfeited_at":     time.Now(),
		"forfeited_by_admin_id": adminID,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update winner status: " + err.Error()})
		return
	}

	// Create new winner record from runner-up
	newWinner := models.Winner{
		DrawID:       runnerUp.DrawID,
		PrizeID:      runnerUp.PrizeID,
		MSISDN:       runnerUp.MSISDN,
		Points:       runnerUp.Points,
		Status:       "pending",
		WinnerNumber: winner.WinnerNumber, // Use same winner number
		IsRunnerUp:   true,
		RunnerUpID:   &runnerUp.ID,
	}
	if err := tx.Create(&newWinner).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new winner from runner-up: " + err.Error()})
		return
	}

	// Update runner-up status to promoted
	if err := tx.Model(&runnerUp).Updates(map[string]interface{}{
		"status":      "promoted",
		"promoted_at": time.Now(),
		"promoted_by_admin_id": adminID,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner-up status: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// Log audit event
	auditEvent := models.AuditLog{
		AdminID:     adminID,
		Action:      "invoke_runner_up",
		EntityType:  "winner",
		EntityID:    winner.ID.String(),
		Description: fmt.Sprintf("Invoked runner-up %s for winner %s. Reason: %s", runnerUp.MSISDN, winner.MSISDN, req.Reason),
	}
	h.auditService.LogAuditEvent(auditEvent)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message":     "Runner-up invoked successfully",
		"old_winner":  winner,
		"runner_up":   runnerUp,
		"new_winner":  newWinner,
	})
}

// Helper function to select winners and runner-ups from eligible participants
func (h *DrawHandler) selectWinnersAndRunnerUps(
	eligibleParticipants []models.ParticipantEvent,
	winnerCount int,
	runnerUpCount int,
	usedMSISDNs map[string]bool,
) ([]models.ParticipantEvent, []models.ParticipantEvent, error) {
	// Create a pool of participants based on points
	var pool []models.ParticipantEvent
	for _, participant := range eligibleParticipants {
		// Skip if this MSISDN has already won
		if usedMSISDNs[participant.MSISDN] {
			continue
		}

		// Add participant to pool based on points
		for i := 0; i < participant.Points; i++ {
			pool = append(pool, participant)
		}
	}

	if len(pool) == 0 {
		return nil, nil, errors.New("no eligible participants available")
	}

	// Shuffle the pool
	// In a real implementation, this would use a secure random number generator
	// For simplicity, we'll use a basic shuffle here
	// rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	// Select winners and runner-ups
	var winners []models.ParticipantEvent
	var runnerUps []models.ParticipantEvent
	selectedMSISDNs := make(map[string]bool)

	// Select winners
	for i := 0; i < winnerCount && len(pool) > 0; i++ {
		// Select a random participant from the pool
		// In a real implementation, this would use a secure random number generator
		// For simplicity, we'll use the first participant in the pool
		winner := pool[0]
		
		// Remove all instances of this MSISDN from the pool
		var newPool []models.ParticipantEvent
		for _, p := range pool {
			if p.MSISDN != winner.MSISDN {
				newPool = append(newPool, p)
			}
		}
		pool = newPool

		winners = append(winners, winner)
		selectedMSISDNs[winner.MSISDN] = true
	}

	// Select runner-ups
	for i := 0; i < runnerUpCount && len(pool) > 0; i++ {
		// Select a random participant from the pool
		// In a real implementation, this would use a secure random number generator
		// For simplicity, we'll use the first participant in the pool
		runnerUp := pool[0]
		
		// Remove all instances of this MSISDN from the pool
		var newPool []models.ParticipantEvent
		for _, p := range pool {
			if p.MSISDN != runnerUp.MSISDN {
				newPool = append(newPool, p)
			}
		}
		pool = newPool

		runnerUps = append(runnerUps, runnerUp)
		selectedMSISDNs[runnerUp.MSISDN] = true
	}

	return winners, runnerUps, nil
}

// ListWinners handles listing all winners
func (h *DrawHandler) ListWinners(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get total count
	var totalCount int64
	if err := h.db.Model(&models.Winner{}).Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count winners: " + err.Error()})
		return
	}

	// Get winners with pagination
	var winners []models.Winner
	if err := h.db.Preload("Draw").
		Preload("Prize").
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&winners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winners: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"winners":     winners,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	})
}

// ExportWinners handles exporting winners to CSV or JSON
func (h *DrawHandler) ExportWinners(c *gin.Context) {
	// Parse query parameters
	format := c.DefaultQuery("format", "json")
	drawIDStr := c.Query("draw_id")

	// Build query
	query := h.db.Model(&models.Winner{}).
		Preload("Draw").
		Preload("Prize")

	// Filter by draw ID if provided
	if drawIDStr != "" {
		drawID, err := uuid.Parse(drawIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw ID format"})
			return
		}
		query = query.Where("draw_id = ?", drawID)
	}

	// Get winners
	var winners []models.Winner
	if err := query.Find(&winners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winners: " + err.Error()})
		return
	}

	// Export based on format
	switch format {
	case "json":
		// Return JSON response
		c.JSON(http.StatusOK, winners)
	case "csv":
		// Generate CSV content
		csvContent := "ID,Draw ID,Draw Date,Prize ID,Prize Name,MSISDN,Points,Status,Created At\n"
		for _, winner := range winners {
			csvContent += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d,%s,%s\n",
				winner.ID,
				winner.DrawID,
				winner.Draw.Date.Format("2006-01-02"),
				winner.PrizeID,
				winner.Prize.Name,
				winner.MSISDN,
				winner.Points,
				winner.Status,
				winner.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}

		// Set headers for CSV download
		c.Header("Content-Disposition", "attachment; filename=winners.csv")
		c.Data(http.StatusOK, "text/csv", []byte(csvContent))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export format. Supported formats: json, csv"})
	}
}

// ClaimPrize handles marking a prize as claimed by a winner
func (h *DrawHandler) ClaimPrize(c *gin.Context) {
	winnerIDStr := c.Param("id")
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid winner ID format"})
		return
	}

	// Get admin ID from context
	adminIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	// Get winner record
	var winner models.Winner
	if err := h.db.First(&winner, winnerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Winner not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winner: " + err.Error()})
		}
		return
	}

	// Check if winner status allows claiming
	if winner.Status != "pending" && winner.Status != "notified" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot claim prize for winner with status: " + winner.Status})
		return
	}

	// Update winner status to claimed
	if err := h.db.Model(&winner).Updates(map[string]interface{}{
		"status":           "claimed",
		"claimed_at":       time.Now(),
		"claimed_by_admin_id": adminID,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update winner status: " + err.Error()})
		return
	}

	// Log audit event
	auditEvent := models.AuditLog{
		AdminID:     adminID,
		Action:      "claim_prize",
		EntityType:  "winner",
		EntityID:    winner.ID.String(),
		Description: fmt.Sprintf("Marked prize as claimed for winner %s", winner.MSISDN),
	}
	h.auditService.LogAuditEvent(auditEvent)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message": "Prize claimed successfully",
		"winner":  winner,
	})
}
