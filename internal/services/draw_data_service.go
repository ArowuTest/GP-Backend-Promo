package services

import (
	"fmt"
	"time"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"gorm.io/gorm"
)

// ParticipantData represents the essential data for a participant in a draw context.
// This is what the DrawHandler will expect from the service.
type ParticipantData struct {
	MSISDN      string
	TotalPoints int
}

// DrawDataService defines the interface for fetching eligible participant data for draws.
// This allows for different implementations (e.g., mock, PostHog, direct DB).
type DrawDataService interface {
	GetEligibleParticipants(drawDate time.Time, prizeStructureID string) ([]ParticipantData, error)
}

// MockDrawDataService is an example implementation for testing or development
type MockDrawDataService struct{}

// GetEligibleParticipants for MockDrawDataService
func (s *MockDrawDataService) GetEligibleParticipants(drawDate time.Time, prizeStructureID string) ([]ParticipantData, error) {
	// Return some mock data
	fmt.Printf("MockDrawDataService: GetEligibleParticipants called for date %s, prizeStructureID %s\n", drawDate.String(), prizeStructureID)
	return []ParticipantData{
		{MSISDN: "2348030000001", TotalPoints: 10},
		{MSISDN: "2348030000002", TotalPoints: 5},
		{MSISDN: "2348030000003", TotalPoints: 20},
		{MSISDN: "2348030000004", TotalPoints: 15},
		{MSISDN: "2348030000005", TotalPoints: 8},
	}, nil
}

// DatabaseDrawDataService implements DrawDataService using direct database queries
type DatabaseDrawDataService struct {
	DB *gorm.DB
}

// NewDatabaseDrawDataService creates a new DatabaseDrawDataService
func NewDatabaseDrawDataService(db *gorm.DB) *DatabaseDrawDataService {
	return &DatabaseDrawDataService{DB: db}
}

// GetEligibleParticipants for DatabaseDrawDataService
// This implementation fetches participant data directly from the database
func (s *DatabaseDrawDataService) GetEligibleParticipants(drawDate time.Time, prizeStructureID string) ([]ParticipantData, error) {
	// Get the prize structure to check applicable days
	var prizeStructure models.PrizeStructure
	if err := s.DB.First(&prizeStructure, "id = ?", prizeStructureID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch prize structure: %w", err)
	}

	// Check if the draw date falls within the prize structure's validity period
	if (prizeStructure.ValidFrom != nil && drawDate.Before(*prizeStructure.ValidFrom)) ||
		(prizeStructure.ValidTo != nil && drawDate.After(*prizeStructure.ValidTo)) {
		return nil, fmt.Errorf("draw date %s is outside the prize structure's validity period", drawDate.Format("2006-01-02"))
	}

	// Check if the day of the week is applicable for this prize structure
	dayOfWeek := drawDate.Weekday().String()[:3] // Get first 3 letters (Mon, Tue, etc.)
	isApplicableDay := false
	for _, day := range prizeStructure.ApplicableDays {
		if day == dayOfWeek {
			isApplicableDay = true
			break
		}
	}

	if !isApplicableDay {
		return nil, fmt.Errorf("draw date %s (%s) is not an applicable day for this prize structure", 
			drawDate.Format("2006-01-02"), dayOfWeek)
	}

	// Set the time bounds for the draw date (start of day to end of day)
	startOfDay := time.Date(drawDate.Year(), drawDate.Month(), drawDate.Day(), 0, 0, 0, 0, drawDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Query participant events for the specified date range
	var participantEvents []models.ParticipantEvent
	if err := s.DB.Where("transaction_timestamp BETWEEN ? AND ?", startOfDay, endOfDay).
		Where("is_eligible = ?", true).
		Find(&participantEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch participant events: %w", err)
	}

	// Check if we have any blacklisted MSISDNs to exclude
	var blacklistedMSISDNs []string
	s.DB.Model(&models.BlacklistedMSISDN{}).Pluck("msisdn", &blacklistedMSISDNs)
	blacklistMap := make(map[string]bool)
	for _, msisdn := range blacklistedMSISDNs {
		blacklistMap[msisdn] = true
	}

	// Aggregate points by MSISDN
	msisdnPoints := make(map[string]int)
	for _, event := range participantEvents {
		// Skip blacklisted MSISDNs
		if blacklistMap[event.MSISDN] {
			continue
		}
		
		// Calculate points based on recharge amount (N100 = 1 point)
		points := int(event.RechargeAmount / 100)
		if points > 0 {
			msisdnPoints[event.MSISDN] += points
		}
	}

	// Convert to ParticipantData slice
	var participants []ParticipantData
	for msisdn, points := range msisdnPoints {
		participants = append(participants, ParticipantData{
			MSISDN:      msisdn,
			TotalPoints: points,
		})
	}

	return participants, nil
}

// PostHogDrawDataService could be an implementation that fetches from PostHog
// This is a placeholder for future implementation
type PostHogDrawDataService struct {
	// PostHog client or configuration
}

// GetEligibleParticipants for PostHogDrawDataService
func (s *PostHogDrawDataService) GetEligibleParticipants(drawDate time.Time, prizeStructureID string) ([]ParticipantData, error) {
	// This would implement logic to query PostHog for the relevant cohort
	// based on the draw date and other criteria
	return nil, fmt.Errorf("PostHogDrawDataService not yet implemented")
}
