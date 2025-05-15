package admin

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv" // Added for StrToInt replacement
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DrawHandler handles draw related requests
type DrawHandler struct {
	DrawDataService services.DrawDataService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(dds services.DrawDataService) *DrawHandler {
	return &DrawHandler{DrawDataService: dds}
}

// ExecuteDrawRequest defines the structure for the execute draw API request
type ExecuteDrawRequest struct {
	DrawDate         string `json:"draw_date" binding:"required"`
	PrizeStructureID string `json:"prize_structure_id" binding:"required"`
}

// ExecuteDrawResponse defines the structure for the execute draw API response
type ExecuteDrawResponse struct {
	DrawID         uuid.UUID             `json:"drawId"`
	DrawDate       time.Time             `json:"drawDate"`
	PrizeStructure models.PrizeStructure `json:"prizeStructure"`
	Winners        []models.DrawWinner   `json:"winners"`
	Status         string                `json:"status"`
	Message        string                `json:"message"`
}

// InvokeRunnerUpRequest defines the structure for invoking a runner-up
type InvokeRunnerUpRequest struct {
	OriginalWinnerID string `json:"originalWinnerId" binding:"required"`
	RunnerUpWinnerID string `json:"runnerUpWinnerId" binding:"required"`
	Notes            string `json:"notes,omitempty"`
}

// csprngIntn returns a cryptographically secure random integer in [0, n)
func csprngIntn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be positive for csprngIntn")
	}
	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(val.Int64()), nil
}

// ExecuteDraw godoc
// @Summary Execute a draw
// @Description Executes a draw for a given date and prize structure
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param drawRequest body ExecuteDrawRequest true "Draw Execution Request"
// @Success 200 {object} ExecuteDrawResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 403 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws/execute [post]
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	var req ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	executedByUserID, err := uuid.Parse(userIDClaim.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw_date format. Expected YYYY-MM-DD"})
		return
	}

	prizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize_structure_id format"})
		return
	}

	var prizeStructure models.PrizeStructure
	if err := config.DB.Preload("Prizes", func(db *gorm.DB) *gorm.DB {
		return db.Order("\"order\" ASC")
	}).First(&prizeStructure, "id = ? AND is_active = ?", prizeStructureID, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active prize structure not found or specified ID is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + err.Error()})
		return
	}

	if (prizeStructure.ValidFrom != nil && drawDate.Before(*prizeStructure.ValidFrom)) ||
		(prizeStructure.ValidTo != nil && drawDate.After(*prizeStructure.ValidTo)) {
		start := "N/A"
		end := "N/A"
		if prizeStructure.ValidFrom != nil {
			start = prizeStructure.ValidFrom.Format("2006-01-02")
		}
		if prizeStructure.ValidTo != nil {
			end = prizeStructure.ValidTo.Format("2006-01-02")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Draw date %s is outside the prize structure's validity period (%s to %s)", drawDate.Format("2006-01-02"), start, end)})
		return
	}

	var existingDraw models.Draw
	if !errors.Is(config.DB.Where("draw_date = ? AND prize_structure_id = ? AND status = ?", drawDate, prizeStructureID, "Completed").First(&existingDraw).Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "A draw for this date and prize structure has already been completed."})
		return
	}

	participants, err := h.DrawDataService.GetEligibleParticipants(drawDate, prizeStructureID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get eligible participants: " + err.Error()})
		return
	}

	if len(participants) == 0 {
		drawAttempt := models.Draw{
			DrawDate:                  drawDate,
			PrizeStructureID:          prizeStructureID,
			Status:                    "Completed - No Participants",
			ExecutedByUserID:          executedByUserID,
			EligibleParticipantsCount: 0,
			TotalPointsInDraw:         0,
		}
		config.DB.Create(&drawAttempt)
		c.JSON(http.StatusOK, ExecuteDrawResponse{
			DrawID:         drawAttempt.ID,
			DrawDate:       drawAttempt.DrawDate,
			PrizeStructure: prizeStructure,
			Winners:        []models.DrawWinner{},
			Status:         drawAttempt.Status,
			Message:        "No eligible participants found for this draw.",
		})
		return
	}

	draw := models.Draw{
		DrawDate:                  drawDate,
		PrizeStructureID:          prizeStructureID,
		Status:                    "Pending",
		ExecutedByUserID:          executedByUserID,
		EligibleParticipantsCount: len(participants),
	}

	var allDrawWinners []models.DrawWinner
	entriesPool := make([]string, 0)
	totalPointsInDraw := 0
	for _, p := range participants {
		for i := 0; i < p.TotalPoints; i++ {
			entriesPool = append(entriesPool, p.MSISDN)
		}
		totalPointsInDraw += p.TotalPoints
	}
	draw.TotalPointsInDraw = totalPointsInDraw

	if len(entriesPool) == 0 {
		draw.Status = "Failed - No Entries"
		config.DB.Create(&draw)
		c.JSON(http.StatusOK, ExecuteDrawResponse{
			DrawID:         draw.ID,
			DrawDate:       draw.DrawDate,
			PrizeStructure: prizeStructure,
			Winners:        []models.DrawWinner{},
			Status:         draw.Status,
			Message:        "No entries available from eligible participants (all had zero points or pool was empty).",
		})
		return
	}

	hasWonInThisDraw := make(map[string]bool)

	for _, prize := range prizeStructure.Prizes {
		numWinnersForPrize := prize.Quantity
		numRunnerUpsForPrize := prize.NumberOfRunnerUps
		if numRunnerUpsForPrize < 0 {
			numRunnerUpsForPrize = 0
		}

		totalSelectionsForPrize := numWinnersForPrize + numRunnerUpsForPrize
		prizeSpecificSelections := []models.DrawWinner{}
		tempPrizePool := make([]string, 0)
		for _, entryMSISDN := range entriesPool {
			if !hasWonInThisDraw[entryMSISDN] {
				tempPrizePool = append(tempPrizePool, entryMSISDN)
			}
		}

		selectedForThisSpecificPrize := make(map[string]bool)

		for k := 0; k < totalSelectionsForPrize; k++ {
			if len(tempPrizePool) == 0 {
				break
			}

			randomIndex, randErr := csprngIntn(len(tempPrizePool))
			if randErr != nil {
				draw.Status = "Failed - RNG Error"
				config.DB.Create(&draw)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate random number: " + randErr.Error()})
				return
			}
			selectedMSISDN := tempPrizePool[randomIndex]

			if selectedForThisSpecificPrize[selectedMSISDN] {
				newTempPool := []string{}
				for _, m := range tempPrizePool {
					if m != selectedMSISDN {
						newTempPool = append(newTempPool, m)
					}
				}
				tempPrizePool = newTempPool
				k-- 
				continue
			}

			dw := models.DrawWinner{
				PrizeID:     prize.ID,
				MSISDN:      selectedMSISDN,
				ClaimStatus: "Pending",
			}

			for _, p := range participants {
				if p.MSISDN == selectedMSISDN {
					dw.PointsAtWin = p.TotalPoints
					break
				}
			}

			if len(prizeSpecificSelections) < numWinnersForPrize {
				dw.IsRunnerUp = false
				hasWonInThisDraw[selectedMSISDN] = true
			} else {
				dw.IsRunnerUp = true
				dw.RunnerUpRank = (len(prizeSpecificSelections) - numWinnersForPrize) + 1
			}
			prizeSpecificSelections = append(prizeSpecificSelections, dw)
			selectedForThisSpecificPrize[selectedMSISDN] = true

			newTempPool := []string{}
			for _, m := range tempPrizePool {
				if m != selectedMSISDN {
					newTempPool = append(newTempPool, m)
				}
			}
			tempPrizePool = newTempPool
		}
		allDrawWinners = append(allDrawWinners, prizeSpecificSelections...)
	}

	draw.Status = "Completed"

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&draw).Error; err != nil {
			return fmt.Errorf("failed to create draw record: %w", err)
		}

		for i := range allDrawWinners {
			allDrawWinners[i].DrawID = draw.ID
		}
		if len(allDrawWinners) > 0 {
			if err := tx.Create(&allDrawWinners).Error; err != nil {
				return fmt.Errorf("failed to save draw winners: %w", err)
			}
		}
		return nil
	})

	if txErr != nil {
		if draw.ID != uuid.Nil {
			config.DB.Model(&models.Draw{}).Where("id = ?", draw.ID).Update("status", "Failed - DB Error")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draw results: " + txErr.Error()})
		return
	}

	var finalDrawResponse models.Draw
	config.DB.Preload("PrizeStructure.Prizes").Preload("Winners.Prize").Preload("ExecutedByUser").First(&finalDrawResponse, draw.ID)

	c.JSON(http.StatusOK, ExecuteDrawResponse{
		DrawID:         finalDrawResponse.ID,
		DrawDate:       finalDrawResponse.DrawDate,
		PrizeStructure: finalDrawResponse.PrizeStructure,
		Winners:        finalDrawResponse.Winners,
		Status:         finalDrawResponse.Status,
		Message:        "Draw executed successfully.",
	})
}

// InvokeRunnerUp godoc
// @Summary Invoke a runner-up
// @Description Promotes a runner-up to a winner if the original winner forfeits
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param invokeRequest body InvokeRunnerUpRequest true "Invoke Runner-Up Request"
// @Success 200 {object} gin.H{"message": string, "updatedWinner": models.DrawWinner, "originalWinnerStatus": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 403 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws/invoke-runner-up [post]
func (h *DrawHandler) InvokeRunnerUp(c *gin.Context) {
	var req InvokeRunnerUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	originalWinnerID, err := uuid.Parse(req.OriginalWinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid original winner ID"})
		return
	}

	runnerUpWinnerID, err := uuid.Parse(req.RunnerUpWinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner-up winner ID"})
		return
	}

	var originalWinner models.DrawWinner
	var runnerUp models.DrawWinner

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Joins("Prize").First(&originalWinner, originalWinnerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("original winner not found")
			}
			return fmt.Errorf("failed to fetch original winner: %w", err)
		}

		if err := tx.Joins("Prize").First(&runnerUp, runnerUpWinnerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("runner-up not found")
			}
			return fmt.Errorf("failed to fetch runner-up: %w", err)
		}

		if !runnerUp.IsRunnerUp {
			return fmt.Errorf("selected user is not a runner-up")
		}
		if originalWinner.IsRunnerUp {
			return fmt.Errorf("cannot invoke a runner-up for another runner-up")
		}
		if originalWinner.DrawID != runnerUp.DrawID || originalWinner.PrizeID != runnerUp.PrizeID {
			return fmt.Errorf("runner-up does not belong to the same prize/draw as the original winner")
		}
		// Use ClaimStatus for checking if the slot is already filled
		if originalWinner.ClaimStatus == "Forfeited-RunnerUpPromoted" {
			return fmt.Errorf("original winner slot has already been filled by a promoted runner-up")
		}
		if runnerUp.ClaimStatus != "Pending" {
			return fmt.Errorf("runner-up is not in a promotable state (current claim status: %s)", runnerUp.ClaimStatus)
		}

		now := time.Now()
		originalWinner.ClaimStatus = "Forfeited-RunnerUpPromoted"
		originalWinner.Notes = fmt.Sprintf("%s; Forfeited on %s, runner-up %s (ID: %s) invoked. %s", originalWinner.Notes, now.Format(time.RFC3339), runnerUp.MSISDN, runnerUp.ID.String(), req.Notes)
		originalWinner.ForfeitedAt = &now
		if err := tx.Save(&originalWinner).Error; err != nil {
			return fmt.Errorf("failed to update original winner: %w", err)
		}

		runnerUp.IsRunnerUp = false
		runnerUp.ClaimStatus = "Pending" // Promoted runner-up starts as Pending claim
		runnerUp.OriginalWinnerID = &originalWinner.ID
		runnerUp.Notes = fmt.Sprintf("Promoted from runner-up (rank %d) for original winner %s (ID: %s). %s", runnerUp.RunnerUpRank, originalWinner.MSISDN, originalWinner.ID.String(), req.Notes)
		runnerUp.RunnerUpRank = 0 // No longer a ranked runner-up
		if err := tx.Save(&runnerUp).Error; err != nil {
			return fmt.Errorf("failed to promote runner-up: %w", err)
		}

		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invoke runner-up: " + txErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Runner-up successfully invoked.",
		"updatedWinner":        runnerUp,
		"originalWinnerStatus": originalWinner.ClaimStatus,
	})
}

// GetDrawDetails godoc
// @Summary Get Draw Details
// @Description Retrieves the details of a specific draw, including winners and prize structure
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param draw_id path string true "Draw ID (UUID)"
// @Success 200 {object} models.Draw
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws/{draw_id} [get]
func (h *DrawHandler) GetDrawDetails(c *gin.Context) {
	drawIDStr := c.Param("draw_id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Draw ID format"})
		return
	}

	var draw models.Draw
	if err := config.DB.Preload("PrizeStructure.Prizes").Preload("Winners.Prize").Preload("ExecutedByUser").First(&draw, drawID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw details: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, draw)
}

// ListDraws godoc
// @Summary List Draws
// @Description Retrieves a paginated list of draws
// @Tags Draws
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number for pagination" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Param sort_by query string false "Field to sort by (e.g., draw_date, status)" default("draw_date")
// @Param sort_order query string false "Sort order (ASC or DESC)" default("DESC")
// @Success 200 {object} gin.H{"draws": []models.Draw, "total": int64, "page": int, "pageSize": int}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/draws [get]
func (h *DrawHandler) ListDraws(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	sortBy := c.DefaultQuery("sort_by", "draw_date")
	sortOrder := c.DefaultQuery("sort_order", "DESC")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var draws []models.Draw
	var total int64

	dbQuery := config.DB.Model(&models.Draw{})

	if err := dbQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count draws: " + err.Error()})
		return
	}

	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	if err := dbQuery.Order(orderClause).Limit(pageSize).Offset(offset).Preload("PrizeStructure").Preload("ExecutedByUser").Find(&draws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draws: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draws":    draws,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

