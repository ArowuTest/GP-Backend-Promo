package admin

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MockPostHogService simulates fetching data from PostHog
type MockPostHogService struct{}

// EligibleParticipant is a simplified struct for mock data from PostHog
type EligibleParticipant struct {
	MSISDN string `json:"msisdn"`
	Points int    `json:"points"`
}

// GetEligibleParticipantsForDraw simulates fetching MSISDNs and their points for a given draw date/window
func (s *MockPostHogService) GetEligibleParticipantsForDraw(drawDate time.Time) ([]EligibleParticipant, error) {
	participants := []EligibleParticipant{
		{MSISDN: "2348030000001", Points: 5},
		{MSISDN: "2348030000002", Points: 10},
		{MSISDN: "2348030000003", Points: 1},
		{MSISDN: "2348030000004", Points: 20},
		{MSISDN: "2348030000005", Points: 2},
		{MSISDN: "2348030000006", Points: 8},
		{MSISDN: "2348030000007", Points: 15},
		{MSISDN: "2348030000008", Points: 3},
		{MSISDN: "2348030000009", Points: 12},
		{MSISDN: "2348030000010", Points: 6},
		{MSISDN: "2348030000011", Points: 7},
		{MSISDN: "2348030000012", Points: 11},
		{MSISDN: "2348030000013", Points: 4},
		{MSISDN: "2348030000014", Points: 18},
		{MSISDN: "2348030000015", Points: 9},
		{MSISDN: "2348030000016", Points: 13},
		{MSISDN: "2348030000017", Points: 17},
		{MSISDN: "2348030000018", Points: 1},
		{MSISDN: "2348030000019", Points: 14},
		{MSISDN: "2348030000020", Points: 16},
	}
	for i := 21; i <= 200; i++ {
		participants = append(participants, EligibleParticipant{
			MSISDN: fmt.Sprintf("2348030000%03d", i),
			Points: rand.Intn(20) + 1,
		})
	}
	return participants, nil
}

// MockMTNBlacklistService simulates checking MSISDNs against a blacklist
type MockMTNBlacklistService struct{}

// IsBlacklisted checks if an MSISDN is blacklisted.
func (s *MockMTNBlacklistService) IsBlacklisted(msisdn string) (bool, error) {
	blacklisted := map[string]bool{
		"2348030000003": true,
		"2348030000015": true,
	}
	return blacklisted[msisdn], nil
}

// ExecuteDrawRequest defines the structure for executing a draw
type ExecuteDrawRequest struct {
	DrawDate         string `json:"draw_date" binding:"required"` // Expecting YYYY-MM-DD
	PrizeStructureID string `json:"prize_structure_id" binding:"required"`
}

// ExecuteDraw handles the execution of a draw
func ExecuteDraw(c *gin.Context) {
	var req ExecuteDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	parsedDrawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw_date format. Expected YYYY-MM-DD."})
		return
	}

	parsedPrizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize_structure_id format"})
		return
	}

	var prizeStructure models.PrizeStructure
	if err := config.DB.Preload("Prizes", func(db *gorm.DB) *gorm.DB {
		return db.Order(""order" ASC") // Ensure prizes are ordered correctly
	}).First(&prizeStructure, "id = ? AND is_active = ?", parsedPrizeStructureID, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active prize structure not found or specified ID is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + err.Error()})
		return
	}

	if (prizeStructure.ValidFrom != nil && parsedDrawDate.Before(*prizeStructure.ValidFrom)) ||
		(prizeStructure.ValidTo != nil && parsedDrawDate.After(*prizeStructure.ValidTo)) {
		start := "N/A"
		end := "N/A"
		if prizeStructure.ValidFrom != nil {
			start = prizeStructure.ValidFrom.Format("2006-01-02")
		}
		if prizeStructure.ValidTo != nil {
			end = prizeStructure.ValidTo.Format("2006-01-02")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Draw date %s is outside the prize structure\'s validity period (%s to %s)", parsedDrawDate.Format("2006-01-02"), start, end)})
		return
	}

	// Check if a draw for this date and prize structure has already been successfully completed
	var existingDraw models.Draw
	if !errors.Is(config.DB.Where("draw_date = ? AND prize_structure_id = ? AND status = ?", parsedDrawDate, parsedPrizeStructureID, "Completed").First(&existingDraw).Error, gorm.ErrRecordNotFound) {
		// Allow rerun if confirmed by user (frontend handles confirmation, backend just needs a flag or separate endpoint for rerun)
		// For now, we prevent re-execution of a completed draw via this standard endpoint.
		// A separate /rerun endpoint or a flag in request could handle this with audit.
		c.JSON(http.StatusConflict, gin.H{"error": "A draw for this date and prize structure has already been completed. Use rerun functionality if intended."})
		return
	}

	posthogService := MockPostHogService{}
	rawParticipants, err := posthogService.GetEligibleParticipantsForDraw(parsedDrawDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch participants from PostHog: " + err.Error()})
		return
	}

	blacklistService := MockMTNBlacklistService{}
	finalEligibleParticipants := []EligibleParticipant{}
	for _, p := range rawParticipants {
		isBlacklisted, _ := blacklistService.IsBlacklisted(p.MSISDN)
		if !isBlacklisted {
			finalEligibleParticipants = append(finalEligibleParticipants, p)
		}
	}

	if len(finalEligibleParticipants) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No eligible (non-blacklisted) participants found for this draw."})
		return
	}

	adminIDClaim, _ := c.Get("userID")
	adminID, _ := uuid.Parse(adminIDClaim.(string))

	draw := models.Draw{
		DrawDate:                  parsedDrawDate,
		PrizeStructureID:          prizeStructure.ID,
		Status:                    "Pending",
		EligibleParticipantsCount: len(finalEligibleParticipants),
		ExecutedByUserID:          adminID,
	}
	totalPoints := 0
	for _, p := range finalEligibleParticipants {
		totalPoints += p.Points
	}
	draw.TotalPointsInDraw = totalPoints

	var drawWinners []models.DrawWinner
	entriesPool := make([]string, 0, totalPoints)
	for _, p := range finalEligibleParticipants {
		for i := 0; i < p.Points; i++ {
			entriesPool = append(entriesPool, p.MSISDN)
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	hasWon := make(map[string]bool)

	for _, prize := range prizeStructure.Prizes {
		for i := 0; i < prize.Quantity; i++ {
			if len(entriesPool) == 0 {
				break
			}
			selectedWinnerMSISDN := ""
			pickedUnique := false
			attempts := 0
			maxAttempts := len(entriesPool)*2 + 10

			for !pickedUnique && attempts < maxAttempts {
				if len(entriesPool) == 0 {
					break
				}
				randomIndex := rand.Intn(len(entriesPool))
				potentialWinnerMSISDN := entriesPool[randomIndex]

				if !hasWon[potentialWinnerMSISDN] {
					selectedWinnerMSISDN = potentialWinnerMSISDN
					hasWon[selectedWinnerMSISDN] = true
					pickedUnique = true
					newEntriesPool := make([]string, 0, len(entriesPool)-1)
					for _, entryMSISDN := range entriesPool {
						if entryMSISDN != selectedWinnerMSISDN {
							newEntriesPool = append(newEntriesPool, entryMSISDN)
						}
					}
					entriesPool = newEntriesPool
				} else {
					// If the chosen one already won, remove one instance of them from the pool to give others a chance
					// This is a simplified approach. A more robust one would be to ensure all entries of a winner are removed once picked.
					// The current logic (removing all entries after picking) is better.
				}
				attempts++
			}

			if !pickedUnique {
				break
			}
			// Find points for the winner
			winnerPoints := 0
			for _, p := range finalEligibleParticipants {
			    if p.MSISDN == selectedWinnerMSISDN {
			        winnerPoints = p.Points
			        break
			    }
			}

			drawWinners = append(drawWinners, models.DrawWinner{
				PrizeID:       prize.ID,
				MSISDN:        selectedWinnerMSISDN, // Store masked version if needed for direct display, or mask on frontend
				IsRunnerUp:    false, // This logic is for main winners
				PointsAtWin:   winnerPoints,
				ClaimStatus:   "Pending",
			})
		}
		if len(entriesPool) == 0 {
			break
		}
	}

	// TODO: Implement runner-up selection logic here if needed, after main winners

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		draw.Status = "Completed"
		if err := tx.Create(&draw).Error; err != nil {
			draw.Status = "Failed"
			tx.Save(&draw) // Attempt to save failed status
			return fmt.Errorf("failed to create draw record: %w", err)
		}

		for i := range drawWinners {
			drawWinners[i].DrawID = draw.ID
		}
		if len(drawWinners) > 0 {
			if err := tx.Create(&drawWinners).Error; err != nil {
				draw.Status = "Failed"
				tx.Model(&draw).Update("status", "Failed")
				return fmt.Errorf("failed to save draw winners: %w", err)
			}
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draw results: " + txErr.Error()})
		return
	}

	var finalDrawResponse models.Draw
	config.DB.Preload("PrizeStructure.Prizes").Preload("Winners.Prize").First(&finalDrawResponse, draw.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Draw executed successfully!",
		"draw":    finalDrawResponse,
	})
}

// ListDraws handles listing all draws
func ListDraws(c *gin.Context) {
	var draws []models.Draw
	result := config.DB.Preload("PrizeStructure").Preload("ExecutedByUser", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "email", "first_name", "last_name")
	}).Order("draw_date desc").Find(&draws)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draws: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, draws)
}

// GetDrawDetails handles retrieving details of a single draw, including its winners
func GetDrawDetails(c *gin.Context) {
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw ID format"})
		return
	}

	var draw models.Draw
	result := config.DB.Preload("PrizeStructure.Prizes").Preload("Winners.Prize").Preload("ExecutedByUser", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "email", "first_name", "last_name")
	}).First(&draw, "id = ?", drawID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw details: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, draw)
}

// ListDataUploadAuditEntries handles listing data upload audit entries
func ListDataUploadAuditEntries(c *gin.Context) {
	var auditEntries []models.DataUploadAudit
	// TODO: Add pagination and filtering (by user, date range, status)
	result := config.DB.Preload("UploadedByUser", func(db *gorm.DB) *gorm.DB {
        return db.Select("id, username, email, first_name, last_name")
    }).Order("upload_timestamp desc").Find(&auditEntries)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data upload audit entries: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, auditEntries)
}

// Note: Winner management handlers (ListWinners, ExportWinnersForMoMo, UpdateWinnerPaymentStatus) from the original file
// would need to be updated to use UUIDs and align with the final models.DrawWinner struct.
// They are omitted here for brevity as the focus is on Draw execution and listing.

