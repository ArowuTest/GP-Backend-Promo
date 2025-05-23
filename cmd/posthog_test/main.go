package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/posthog"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	// Get PostHog credentials from environment
	apiKey := os.Getenv("POSTHOG_API_KEY")
	projectID := os.Getenv("POSTHOG_PROJECT_ID")
	baseURL := os.Getenv("POSTHOG_BASE_URL")

	if apiKey == "" || projectID == "" || baseURL == "" {
		log.Fatal("Missing required PostHog environment variables")
	}

	// Create PostHog client
	client := posthog.NewClient(apiKey, projectID, baseURL)
	cohortGenerator := posthog.NewCohortGenerator(client)
	dataSource := posthog.NewPostHogDataSource(client, cohortGenerator)

	// Create context
	ctx := context.Background()

	// Test cohort creation
	fmt.Println("Creating test cohort...")
	cohortID, err := cohortGenerator.CreateTestCohort(ctx)
	if err != nil {
		log.Fatalf("Failed to create test cohort: %v", err)
	}
	fmt.Printf("Created test cohort with ID: %s\n", cohortID)

	// Test getting eligible participants
	fmt.Println("Getting eligible participants for today...")
	participants, err := dataSource.GetEligibleParticipants(ctx, time.Now())
	if err != nil {
		log.Fatalf("Failed to get eligible participants: %v", err)
	}
	fmt.Printf("Found %d eligible participants\n", len(participants))

	// Print first few participants if available
	if len(participants) > 0 {
		fmt.Println("Sample participants:")
		limit := 5
		if len(participants) < limit {
			limit = len(participants)
		}
		for i := 0; i < limit; i++ {
			p := participants[i]
			fmt.Printf("  MSISDN: %s, Points: %d, Recharge Amount: %.2f\n", 
				maskMSISDN(p.MSISDN), p.Points, p.RechargeAmount)
		}
	}

	fmt.Println("PostHog integration test completed successfully")
}

// maskMSISDN masks the middle digits of an MSISDN for privacy
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 6 {
		return msisdn
	}
	
	prefix := msisdn[:3]
	suffix := msisdn[len(msisdn)-3:]
	masked := prefix + "****" + suffix
	
	return masked
}
