package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ArowuTest/GP-Backend-Promo/internal/api"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
)

func main() {
	// V5 DEPLOYMENT DIAGNOSTIC
	fmt.Fprintf(os.Stdout, ">>>>> V5 DEPLOYMENT PIPELINE TEST - MAIN.GO STARTED SUCCESSFULLY (STDOUT) <<<<<\n")
	fmt.Fprintf(os.Stderr, ">>>>> V5 DEPLOYMENT PIPELINE TEST - MAIN.GO STARTED SUCCESSFULLY (STDERR) <<<<<\n")
	log.Println(">>>>> V5 DEPLOYMENT PIPELINE TEST - MAIN.GO STARTED SUCCESSFULLY (LOG) <<<<<")


	// Load environment variables and connect to the database
	config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set
	}

	// Initialize Gin router
	router := api.SetupRouter()

	log.Printf("Backend server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

