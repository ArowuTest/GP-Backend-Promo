package admin

import (
	"fmt"
	"math/rand"
	"net/http"
	// "sort" // Unused import removed
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mynumba_don_win_draw_system_backend_internal_config "mynumba-don-win-draw-system/backend/internal/config"
	mynumba_don_win_draw_system_backend_internal_models "mynumba-don-win-draw-system/backend/internal/models"
	"gorm.io/gorm"
)

// MockPostHogService simulates fetching data from PostHog
type MockPostHogService struct{}

// GetEligibleParticipantsForDraw simulates fetching MSISDNs and their points for a given draw date/window
// In a real scenario, this would query PostHog segments.
func (s *MockPostHogService) GetEligibleParticipantsForDraw(drawDate time.Time) ([]mynumba_don_win_draw_system_backend_internal_models.EligibleParticipant, error) {
	// For now, using the sample CSV data logic. This needs to be replaced with actual PostHog API calls.
	// This is a placeholder. We will need to read and parse the CSV file provided by the user.
	// For simplicity, let's return a fixed list for now.
	// The actual implementation will involve: 
	// 1. Defining how to query PostHog (API client, specific segments based on drawDate)
	// 2. Mapping PostHog response to []EligibleParticipant
	
	// Simulate a few participants with varying points
	// In a real system, MSISDNs would be actual phone numbers.
	// Points are derived from recharge amounts (e.g., N100 = 1 point)
	participants := []mynumba_don_win_draw_system_backend_internal_models.EligibleParticipant{
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
        // Add more to ensure enough for prize tiers
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

    // Simulate enough participants for a large draw
    for i := 21; i <= 200; i++ {
        participants = append(participants, mynumba_don_win_draw_system_backend_internal_models.EligibleParticipant{
            MSISDN: fmt.Sprintf("2348030000%03d", i),
            Points: rand.Intn(20) + 1, // Random points between 1 and 20
        })
    }

	return participants, nil
}

// MockMTNBlacklistService simulates checking MSISDNs against a blacklist
type MockMTNBlacklistService struct{}

// IsBlacklisted checks if an MSISDN is blacklisted.
// In a real scenario, this would call the MTN Blacklist API.
func (s *MockMTNBlacklistService) IsBlacklisted(msisdn string) (bool, error) {
	// Simulate a few blacklisted numbers for testing
	blacklisted := map[string]bool{
		"2348030000003": true, // Example blacklisted number
		"2348030000015": true, // Another example
	}
	return blacklisted[msisdn], nil
}

// ExecuteDrawRequest defines the structure for executing a draw
type ExecuteDrawRequest struct {
	DrawDate         string `json:"drawDate" binding:"required"` // Expecting YYYY-MM-DD
	PrizeStructureID string `json:"prizeStructureID" binding:"required"`
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawDate format. Expected YYYY-MM-DD."})
        return
    }

	parsedPrizeStructureID, err := uuid.Parse(req.PrizeStructureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	// 1. Fetch the active Prize Structure
	var prizeStructure mynumba_don_win_draw_system_backend_internal_models.PrizeStructure
	if err := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeTiers", func(db *gorm.DB) *gorm.DB {
        return db.Order("prize_tiers.sort_order ASC") // Ensure tiers are ordered correctly by sort_order
    }).Where("id = ? AND is_active = ?", parsedPrizeStructureID, true).First(&prizeStructure).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active prize structure not found or specified ID is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + err.Error()})
		return
	}
    // Validate if the drawDate falls within the prize structure's validity period
    if (prizeStructure.EffectiveStartDate != nil && parsedDrawDate.Before(*prizeStructure.EffectiveStartDate)) || 
       (prizeStructure.EffectiveEndDate != nil && parsedDrawDate.After(*prizeStructure.EffectiveEndDate)) {
        
		start := "N/A"
		end := "N/A"
		if prizeStructure.EffectiveStartDate != nil {
			start = prizeStructure.EffectiveStartDate.Format("2006-01-02")
		}
		if prizeStructure.EffectiveEndDate != nil {
			end = prizeStructure.EffectiveEndDate.Format("2006-01-02")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Draw date %s is outside the prize structure's validity period (%s to %s)", parsedDrawDate.Format("2006-01-02"), start, end)})
        return
    }


	// 2. Fetch eligible participants from PostHog (mocked for now)
	posthogService := MockPostHogService{}
	rawParticipants, err := posthogService.GetEligibleParticipantsForDraw(parsedDrawDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch participants from PostHog: " + err.Error()})
		return
	}

	// 3. Filter out blacklisted participants (mocked MTN Blacklist API)
	blacklistService := MockMTNBlacklistService{}
	finalEligibleParticipants := []mynumba_don_win_draw_system_backend_internal_models.EligibleParticipant{}
	for _, p := range rawParticipants {
		isBlacklisted, err := blacklistService.IsBlacklisted(p.MSISDN)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking blacklist for " + p.MSISDN + ": " + err.Error()})
			return
		}
		if !isBlacklisted {
			finalEligibleParticipants = append(finalEligibleParticipants, p)
		}
	}

	if len(finalEligibleParticipants) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No eligible (non-blacklisted) participants found for this draw."})
		return
	}

	// 4. Create the draw entry in the database
	adminIDClaim, _ := c.Get("userID") // Assuming userID is string from JWT, convert to UUID
    adminID, _ := uuid.Parse(adminIDClaim.(string))

	draw := mynumba_don_win_draw_system_backend_internal_models.Draw{
		DrawDate:                parsedDrawDate,
		PrizeStructureID:      prizeStructure.ID,
		Status:                  mynumba_don_win_draw_system_backend_internal_models.DrawStatusPending, // Will update to Completed or Failed
		TotalEligibleMSISDNs:  func(i int) *int { return &i }(len(finalEligibleParticipants)),
		ExecutedByAdminID:     &adminID,
		ExecutionType:         mynumba_don_win_draw_system_backend_internal_models.ExecutionManual, // Assuming manual for now
	}
	totalTickets := 0
	for _, p := range finalEligibleParticipants {
		totalTickets += p.Points
	}
	draw.TotalTickets = &totalTickets

	// 5. Perform the draw - Select winners
	var winners []mynumba_don_win_draw_system_backend_internal_models.Winner
	entriesPool := []string{} // Each MSISDN is added Points times
	for _, p := range finalEligibleParticipants {
		for i := 0; i < p.Points; i++ {
			entriesPool = append(entriesPool, p.MSISDN)
		}
	}

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	// Create a map to track who has already won to ensure unique winners per draw
	hasWon := make(map[string]bool)

	for _, tier := range prizeStructure.PrizeTiers { // Corrected: removed unused tierIdx
		for i := 0; i < tier.WinnerCount; i++ {
			if len(entriesPool) == 0 {
				// Not enough unique participants for all prize slots
				break // Break from this tier's winner selection
			}

			// Try to pick a unique winner
			var selectedWinnerMSISDN string
			pickedUnique := false
			attempts := 0 // To prevent infinite loops if all remaining in pool have won
			maxAttempts := len(entriesPool) * 2 + 10 // Heuristic for max attempts

			for !pickedUnique && attempts < maxAttempts {
				if len(entriesPool) == 0 { break }
				randomIndex := rand.Intn(len(entriesPool))
				potentialWinnerMSISDN := entriesPool[randomIndex]

				if !hasWon[potentialWinnerMSISDN] {
					selectedWinnerMSISDN = potentialWinnerMSISDN
					hasWon[selectedWinnerMSISDN] = true
					pickedUnique = true
					// Remove all entries of this winner from the pool to ensure they don't win again
					newEntriesPool := []string{}
					for _, entryMSISDN := range entriesPool {
						if entryMSISDN != selectedWinnerMSISDN {
							newEntriesPool = append(newEntriesPool, entryMSISDN)
						}
					}
					entriesPool = newEntriesPool
				} else {
					// This MSISDN already won, try picking another. 
					// (If they were removed from pool, this case is less likely for subsequent picks in same tier, but good for safety)
				}
				attempts++
			}

			if !pickedUnique {
				// Could not find a unique winner for this slot (e.g., all remaining participants already won)
				break // Break from this tier's winner selection
			}
			selectionOrder := i + 1
			winners = append(winners, mynumba_don_win_draw_system_backend_internal_models.Winner{
				MSISDN:      selectedWinnerMSISDN,
				PrizeTierID: tier.ID,
				PrizeAmountWon: tier.PrizeAmount,
				SelectionOrderInTier: &selectionOrder,
				NotificationStatus: mynumba_don_win_draw_system_backend_internal_models.NotificationPending, // Initial status
				PaymentStatus: mynumba_don_win_draw_system_backend_internal_models.PaymentPendingExport, // Initial status
			})
		}
		// Corrected: The undefined 'i' error was likely a misinterpretation by the compiler or a subtle issue.
		// The original logic for 'i' in the loop and the break condition seems correct for its scope.
		// If the pool is empty and we haven't selected all winners for *this tier* (i < tier.WinnerCount-1),
		// then we should break the outer loop over tiers.
		// However, the condition `i < tier.WinnerCount-1` might be problematic if `i` is the loop counter for winners in the current tier.
		// Let's re-evaluate the break condition. We should break the outer loop if `entriesPool` is empty
		// and we still have tiers to process or winners to pick in the current tier.
		// A simpler break: if entriesPool is empty after trying to pick winners for a tier, we can't pick more for subsequent tiers.
		if len(entriesPool) == 0 {
            break // Break outer loop (over tiers) if no more participants to pick from
        }
	}

	// 6. Save Draw and Winners in a transaction
	txErr := mynumba_don_win_draw_system_backend_internal_config.DB.Transaction(func(tx *gorm.DB) error {
		draw.Status = mynumba_don_win_draw_system_backend_internal_models.DrawStatusCompleted
		if err := tx.Create(&draw).Error; err != nil {
			draw.Status = mynumba_don_win_draw_system_backend_internal_models.DrawStatusFailed
			// Attempt to save failed draw status, but prioritize original error
			tx.Save(&draw) // Save the draw even if it failed, with status Failed
			return err
		}

		for idx := range winners { // Use idx to avoid shadowing outer loop variables if any
			winners[idx].DrawID = draw.ID
		}
		if len(winners) > 0 {
		    if err := tx.Create(&winners).Error; err != nil {
		        draw.Status = mynumba_don_win_draw_system_backend_internal_models.DrawStatusFailed // Mark draw as failed if winners can't be saved
		        tx.Save(&draw)
		        return err
		    }
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draw results: " + txErr.Error()})
		return
	}

	// 7. Prepare and return response
	// Fetch the draw again with its winners and their prize tier details for a comprehensive response
    var finalDrawResponse mynumba_don_win_draw_system_backend_internal_models.Draw
    mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeStructure.PrizeTiers").Preload("Winners.PrizeTier").First(&finalDrawResponse, draw.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Draw executed successfully!",
		"draw":    finalDrawResponse,
	})
}

// ListDraws handles listing all draws
func ListDraws(c *gin.Context) {
	var draws []mynumba_don_win_draw_system_backend_internal_models.Draw
	// Add pagination, filtering by date range, status etc. later
	result := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeStructure").Preload("ExecutedByAdminID", func(db *gorm.DB) *gorm.DB {
        return db.Select("id, email, first_name, last_name") // Select only necessary fields
    }).Order("draw_date desc").Find(&draws)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draws: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, draws)
}

// GetDrawDetails handles retrieving details of a single draw, including its winners
func GetDrawDetails(c *gin.Context) {
	drawID := c.Param("id")
	parsedDrawID, err := uuid.Parse(drawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw ID format"})
		return
	}

	var draw mynumba_don_win_draw_system_backend_internal_models.Draw
	result := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeStructure.PrizeTiers").Preload("Winners.PrizeTier").Preload("ExecutedByAdminID", func(db *gorm.DB) *gorm.DB {
        return db.Select("id, email, first_name, last_name")
    }).Where("id = ?", parsedDrawID).First(&draw)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw details: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, draw)
}

// RerunDraw - Placeholder for now. This would be a complex operation.
// Considerations: 
// - Why is a rerun needed? (e.g., technical issue, incorrect setup)
// - Should previous winners be invalidated?
// - Audit trail for reruns.
// - Potential impact on already notified winners.
// For now, this is out of scope for initial implementation based on FSD.
func RerunDraw(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Rerun draw functionality is not yet implemented."})
}

// --- Winner Management Handlers (Placeholders) ---

// ListWinners handles listing winners with filtering options
func ListWinners(c *gin.Context) {
    // TODO: Implement filtering (by draw date, prize tier, MSISDN, payment status)
    // TODO: Implement pagination
    var winners []mynumba_don_win_draw_system_backend_internal_models.Winner
    result := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("Draw.PrizeStructure").Preload("PrizeTier").Order("created_at DESC").Find(&winners)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve winners: " + result.Error.Error()})
        return
    }
    c.JSON(http.StatusOK, winners)
}

// ExportWinnersForMoMo handles exporting winners in the specified MoMo format
func ExportWinnersForMoMo(c *gin.Context) {
    // TODO: 
    // 1. Fetch winners (potentially filtered by draw ID or date range passed in query params)
    // 2. Format data according to: Winner MSISDN, Prize Amount, Date of Draw, and Draw Prize Position.
    // 3. Generate CSV file content
    // 4. Set appropriate headers for CSV download
    // 5. Update winner payment status to PaymentExportedForPayment
    c.JSON(http.StatusNotImplemented, gin.H{"message": "Export winners for MoMo functionality is not yet implemented."})
}

// UpdateWinnerPaymentStatus handles updating the payment status of a winner
func UpdateWinnerPaymentStatus(c *gin.Context) {
    winnerID := c.Param("id")
    parsedWinnerID, err := uuid.Parse(winnerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid winner ID format"})
        return
    }

    var req struct {
        Status mynumba_don_win_draw_system_backend_internal_models.PaymentStatus `json:"status" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
        return
    }

    // Validate status value (add more specific validation if needed)
    if req.Status != mynumba_don_win_draw_system_backend_internal_models.PaymentConfirmed && 
       req.Status != mynumba_don_win_draw_system_backend_internal_models.PaymentFailed && 
       req.Status != mynumba_don_win_draw_system_backend_internal_models.PaymentPendingExport && 
       req.Status != mynumba_don_win_draw_system_backend_internal_models.PaymentExportedForPayment &&
       req.Status != mynumba_don_win_draw_system_backend_internal_models.PaymentRequiresVerification {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment status value"})
        return
    }

    var winner mynumba_don_win_draw_system_backend_internal_models.Winner
    if mynumba_don_win_draw_system_backend_internal_config.DB.Where("id = ?", parsedWinnerID).First(&winner).Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Winner not found"})
        return
    }

    result := mynumba_don_win_draw_system_backend_internal_config.DB.Model(&winner).Update("payment_status", req.Status)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update winner payment status: " + result.Error.Error()})
        return
    }

    mynumba_don_win_draw_system_backend_internal_config.DB.Preload("Draw.PrizeStructure").Preload("PrizeTier").First(&winner, "id = ?", parsedWinnerID)
    c.JSON(http.StatusOK, winner)
}

