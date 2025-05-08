package main

import (
	"log"
	"os"

	mynumba_don_win_draw_system_backend_internal_api "mynumba-don-win-draw-system/backend/internal/api"
	mynumba_don_win_draw_system_backend_internal_config "mynumba-don-win-draw-system/backend/internal/config"
)

func main() {
	// Load environment variables and connect to the database
	mynumba_don_win_draw_system_backend_internal_config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Initialize Gin router
	router := mynumba_don_win_draw_system_backend_internal_api.SetupRouter()

	log.Printf("Backend server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

