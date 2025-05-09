package main

import (
	"log"
	"os"

	"github.com/ArowuTest/GP-Backend-Promo/internal/api"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
)

func main() {
	// Load environment variables and connect to the database
	config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Gin router
	router := api.SetupRouter()

	log.Printf("Backend server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

