package draw

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// ExecuteDrawService provides functionality for executing draws
type ExecuteDrawService struct {
	drawRepository        draw.DrawRepository
	participantRepository participant.ParticipantRepository
	prizeRepository       prize.PrizeRepository
	auditService          audit.AuditService
}

// NewDrawService creates a new ExecuteDrawService
func NewDrawService(
	drawRepository draw.DrawRepository,
	participantRepository participant.ParticipantRepository,
	prizeRepository prize.PrizeRepository,
	auditService audit.AuditService,
) *ExecuteDrawService {
	return &ExecuteDrawService{
		drawRepository:        drawRepository,
		participantRepository: participantRepository,
		prizeRepository:       prizeRepository,
		auditService:          auditService,
	}
}

// ExecuteDrawInput defines the input for the ExecuteDraw use case
type ExecuteDrawInput struct {
	DrawDate         time.Time
	PrizeStructureID uuid.UUID
	ExecutedByAdminID uuid.UUID
}

// ExecuteDrawOutput defines the output for the ExecuteDraw use case
type ExecuteDrawOutput struct {
	DrawID              uuid.UUID
	DrawDate            time.Time
	TotalEligibleMSISDNs int
	TotalEntries        int
	Winners             []WinnerOutput
}

// WinnerOutput defines the winner output structure
type WinnerOutput struct {
	ID          uuid.UUID
	MSISDN      string
	PrizeTierID uuid.UUID
	PrizeName   string
	PrizeValue  string
}

// ExecuteDraw executes a draw for the given date and prize structure
func (uc *ExecuteDrawService) ExecuteDraw(input ExecuteDrawInput) (*ExecuteDrawOutput, error) {
	// Validate input
	if input.DrawDate.IsZero() {
		return nil, errors.New("draw date is required")
	}
	
	if input.PrizeStructureID == uuid.Nil {
		return nil, errors.New("prize structure ID is required")
	}
	
	if input.ExecutedByAdminID == uuid.Nil {
		return nil, errors.New("executed by admin ID is required")
	}
	
	// Check if draw already exists for the date
	existingDraw, err := uc.drawRepository.GetByDate(input.DrawDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing draw: %w", err)
	}
	
	if existingDraw != nil {
		return nil, draw.NewDrawError(draw.ErrDrawAlreadyExists, "Draw already exists for this date", nil)
	}
	
	// Get prize structure
	prizeStructure, err := uc.prizeRepository.GetPrizeStructureByID(input.PrizeStructureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prize structure: %w", err)
	}
	
	// Get eligible participants
	eligibleParticipants, err := uc.getEligibleParticipants(input.DrawDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible participants: %w", err)
	}
	
	if len(eligibleParticipants) == 0 {
		return nil, draw.NewDrawError(draw.ErrNoEligibleParticipants, "No eligible participants for draw", nil)
	}
	
	// Calculate total entries
	totalEntries := 0
	for _, participant := range eligibleParticipants {
		totalEntries += participant.Points
	}
	
	// Create draw
	drawID := uuid.New()
	newDraw := &draw.Draw{
		ID:                   drawID,
		DrawDate:             input.DrawDate,
		PrizeStructureID:     input.PrizeStructureID,
		Status:               "Pending",
		TotalEligibleMSISDNs: len(eligibleParticipants),
		TotalEntries:         totalEntries,
		ExecutedByAdminID:    input.ExecutedByAdminID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	
	if err := uc.drawRepository.Create(newDraw); err != nil {
		return nil, fmt.Errorf("failed to create draw: %w", err)
	}
	
	// Execute draw algorithm
	winners, err := uc.executeDrawAlgorithm(newDraw, prizeStructure, eligibleParticipants)
	if err != nil {
		// Update draw status to failed
		newDraw.Status = "Failed"
		if updateErr := uc.drawRepository.Update(newDraw); updateErr != nil {
			// Log error but continue with original error
			fmt.Printf("Failed to update draw status: %v\n", updateErr)
		}
		
		return nil, fmt.Errorf("failed to execute draw algorithm: %w", err)
	}
	
	// Update draw status to completed
	newDraw.Status = "Completed"
	newDraw.Winners = winners
	if err := uc.drawRepository.Update(newDraw); err != nil {
		return nil, fmt.Errorf("failed to update draw status: %w", err)
	}
	
	// Create winners
	winnerOutputs := make([]WinnerOutput, 0, len(winners))
	for _, winner := range winners {
		if err := uc.drawRepository.CreateWinner(&winner); err != nil {
			return nil, fmt.Errorf("failed to create winner: %w", err)
		}
		
		// Find prize tier
		var prizeName, prizeValue string
		for _, prizeTier := range prizeStructure.Prizes {
			if prizeTier.ID == winner.PrizeTierID {
				prizeName = prizeTier.Name
				prizeValue = prizeTier.Value
				break
			}
		}
		
		winnerOutputs = append(winnerOutputs, WinnerOutput{
			ID:          winner.ID,
			MSISDN:      winner.MSISDN,
			PrizeTierID: winner.PrizeTierID,
			PrizeName:   prizeName,
			PrizeValue:  prizeValue,
		})
	}
	
	// Log audit
	if err := uc.auditService.LogAudit(
		"EXECUTE_DRAW",
		"Draw",
		drawID,
		input.ExecutedByAdminID,
		fmt.Sprintf("Draw executed for date %s", input.DrawDate.Format("2006-01-02")),
		fmt.Sprintf("Total eligible MSISDNs: %d, Total entries: %d, Winners: %d", 
			len(eligibleParticipants), totalEntries, len(winners)),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &ExecuteDrawOutput{
		DrawID:              drawID,
		DrawDate:            input.DrawDate,
		TotalEligibleMSISDNs: len(eligibleParticipants),
		TotalEntries:        totalEntries,
		Winners:             winnerOutputs,
	}, nil
}

	// getEligibleParticipants retrieves eligible participants for the draw
func (uc *ExecuteDrawService) getEligibleParticipants(date time.Time) ([]participant.Participant, error) {
	// This is a simplified implementation
	// In a real-world scenario, this would involve complex eligibility rules
	
	// Get all participants with points up to the draw date
	participants, _, err := uc.participantRepository.ListByDate(date, 1, 1000)
	if err != nil {
		return nil, err
	}
	
	return participants, nil
}
// executeDrawAlgorithm implements the draw algorithm
func (uc *ExecuteDrawService) executeDrawAlgorithm(
	newDraw *draw.Draw,
	prizeStructure *prize.PrizeStructure,
	eligibleParticipants []participant.Participant,
) ([]draw.Winner, error) {
	// This is a simplified implementation of the draw algorithm
	// In a real-world scenario, this would be more complex
	
	// Create a pool of entries based on points
	entryPool := make([]string, 0, newDraw.TotalEntries)
	for _, participant := range eligibleParticipants {
		for i := 0; i < participant.Points; i++ {
			entryPool = append(entryPool, participant.MSISDN)
		}
	}
	
	// Shuffle the entry pool
	shuffleEntries(entryPool)
	
	// Select winners for each prize tier
	winners := make([]draw.Winner, 0)
	selectedMSISDNs := make(map[string]bool)
	
	for _, prizeTier := range prizeStructure.Prizes {
		// For each prize in the tier
		for i := 0; i < prizeTier.Quantity; i++ {
			// Find a winner that hasn't been selected yet
			var winnerMSISDN string
			for len(entryPool) > 0 {
				// Pop an entry from the pool
				winnerMSISDN = entryPool[0]
				entryPool = entryPool[1:]
				
				// Check if this MSISDN has already won
				if !selectedMSISDNs[winnerMSISDN] {
					selectedMSISDNs[winnerMSISDN] = true
					break
				}
			}
			
			if winnerMSISDN == "" {
				return nil, errors.New("not enough eligible participants for all prizes")
			}
			
			// Create winner
			winner := draw.Winner{
				ID:            uuid.New(),
				DrawID:        newDraw.ID,
				MSISDN:        winnerMSISDN,
				PrizeTierID:   prizeTier.ID,
				Status:        "PendingNotification",
				PaymentStatus: "Pending",
				IsRunnerUp:    false,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			
			winners = append(winners, winner)
		}
		
		// Generate runner-ups for this prize tier (50% of winners or at least 1)
		runnerUpCount := prizeTier.Quantity / 2
		if runnerUpCount < 1 {
			runnerUpCount = 1
		}
		
		for i := 0; i < runnerUpCount; i++ {
			// Find a runner-up that hasn't been selected yet
			var runnerUpMSISDN string
			for len(entryPool) > 0 {
				// Pop an entry from the pool
				runnerUpMSISDN = entryPool[0]
				entryPool = entryPool[1:]
				
				// Check if this MSISDN has already won
				if !selectedMSISDNs[runnerUpMSISDN] {
					selectedMSISDNs[runnerUpMSISDN] = true
					break
				}
			}
			
			if runnerUpMSISDN == "" {
				// Not enough participants for runner-ups, but that's okay
				break
			}
			
			// Create runner-up
			runnerUp := draw.Winner{
				ID:            uuid.New(),
				DrawID:        newDraw.ID,
				MSISDN:        runnerUpMSISDN,
				PrizeTierID:   prizeTier.ID,
				Status:        "PendingNotification",
				PaymentStatus: "Pending",
				IsRunnerUp:    true,
				RunnerUpRank:  i + 1,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			
			winners = append(winners, runnerUp)
		}
	}
	
	return winners, nil
}

// shuffleEntries shuffles the entry pool
func shuffleEntries(entries []string) {
	// Fisher-Yates shuffle algorithm
	for i := len(entries) - 1; i > 0; i-- {
		j := uuid.New().ID() % uint32(i+1)
		entries[i], entries[j] = entries[j], entries[i]
	}
}
