package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin"
	"github.com/ArowuTest/GP-Backend-Promo/internal/middleware"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	// Initialize database connection
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set up Gin router
	router := setupRouter()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Handle OPTIONS requests
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		api.POST("/auth/login", admin.Login)

		// Admin routes (auth required)
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.JWTAuth())
		{
			// User management
			adminRoutes.GET("/users", admin.ListUsers)
			adminRoutes.POST("/users", middleware.RequireRole("SUPER_ADMIN"), admin.CreateUser)
			adminRoutes.GET("/users/:id", admin.GetUser)
			adminRoutes.PUT("/users/:id", middleware.RequireRole("SUPER_ADMIN"), admin.UpdateUser)
			adminRoutes.DELETE("/users/:id", middleware.RequireRole("SUPER_ADMIN"), admin.DeleteUser)

			// Prize structure management
			prizeStructureHandler := admin.NewPrizeStructureHandler()
			adminRoutes.GET("/prize-structures/", prizeStructureHandler.ListPrizeStructures)
			adminRoutes.POST("/prize-structures/", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), prizeStructureHandler.CreatePrizeStructure)
			adminRoutes.GET("/prize-structures/:id", prizeStructureHandler.GetPrizeStructure)
			adminRoutes.PUT("/prize-structures/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), prizeStructureHandler.UpdatePrizeStructure)
			adminRoutes.DELETE("/prize-structures/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), prizeStructureHandler.DeletePrizeStructure)

			// Participant data management
			participantHandler := admin.NewParticipantHandler()
			adminRoutes.POST("/participants/upload", middleware.RequireRole("SUPER_ADMIN", "ADMIN", "SENIOR_USER"), participantHandler.UploadParticipants)
			adminRoutes.GET("/participants", participantHandler.ListParticipants)

			// Draw management
			// Use the real database service instead of mock
			drawService := services.NewDatabaseDrawDataService(config.DB)
			drawHandler := admin.NewDrawHandler(drawService)
			adminRoutes.POST("/draws/execute", middleware.RequireRole("SUPER_ADMIN"), drawHandler.ExecuteDraw)
			adminRoutes.GET("/draws/eligibility-stats", drawHandler.GetEligibilityStats)
			adminRoutes.GET("/draws", drawHandler.ListDraws)
			adminRoutes.GET("/draws/:draw_id", drawHandler.GetDrawDetails)
			adminRoutes.POST("/draws/invoke-runner-up", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), drawHandler.InvokeRunnerUp)

			// Reports
			adminRoutes.GET("/reports/data-uploads/", admin.ListDataUploadAuditEntries)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
	})

	// Handle 404 Not Found
	router.NoRoute(func(c *gin.Context) {
		// Check if the request is for the API
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("API endpoint not found: %s", c.Request.URL.Path)})
			return
		}
		// For non-API routes, return a simple 404 message
		c.String(http.StatusNotFound, "Page not found")
	})

	return router
}
