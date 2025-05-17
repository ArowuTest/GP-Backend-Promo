package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers"
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin"
	"github.com/ArowuTest/GP-Backend-Promo/internal/middleware"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AppHandlers holds all handler instances
type AppHandlers struct {
	AdminUserHandler        *handlers.AdminUserHandler
	DataUploadAuditHandler  *handlers.DataUploadAuditHandler
	AdminHandlers           *admin.AdminHandlers
	// Add other handlers here as needed
}

// AppServices holds all service instances
type AppServices struct {
	AuditService    *services.AuditService
	DrawDataService services.DrawDataService
	// Add other services here as needed
}

func main() {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Initialize database connection
	db, err := initializeDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize services with DB dependency injection
	appServices := initializeServices(db)

	// Initialize handlers with services and DB dependency injection
	appHandlers := initializeHandlers(db, appServices)

	// Setup router with handlers
	router := setupRouter(appHandlers)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeDB sets up the database connection
func initializeDB() (*gorm.DB, error) {
	// This function should initialize and return the DB connection
	// For now, we'll use the existing DB initialization from config
	// In the future, this could be refactored to not use a global DB variable
	return config.InitDB()
}

// initializeServices initializes all services with DB dependency injection
func initializeServices(db *gorm.DB) *AppServices {
	return &AppServices{
		AuditService:    services.NewAuditService(db),
		DrawDataService: services.NewDatabaseDrawDataService(db),
		// Initialize other services here
	}
}

// initializeHandlers initializes all handlers with service dependencies
func initializeHandlers(db *gorm.DB, services *AppServices) *AppHandlers {
	return &AppHandlers{
		AdminUserHandler:       handlers.NewAdminUserHandler(db),
		DataUploadAuditHandler: handlers.NewDataUploadAuditHandler(db),
		AdminHandlers: admin.NewAdminHandlers(
			services.AuditService,
			services.DrawDataService,
			// Pass other required services here
		),
		// Initialize other handlers here
	}
}

// setupRouter configures the HTTP router with all routes and middleware
func setupRouter(h *AppHandlers) *gin.Engine {
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
		c.Status(http.StatusOK)
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/health", healthCheck)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.AdminUserHandler.Login)
			// Add other auth routes
		}

		// Admin routes (protected)
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth())
		{
			// User management routes
			admin.GET("/users", h.AdminUserHandler.ListUsers)
			admin.POST("/users", h.AdminUserHandler.CreateUser)
			admin.GET("/users/:id", h.AdminUserHandler.GetUser)
			admin.PUT("/users/:id", h.AdminUserHandler.UpdateUser)
			admin.DELETE("/users/:id", h.AdminUserHandler.DeleteUser)

			// Data upload audit routes
			admin.POST("/audits/data-uploads", h.DataUploadAuditHandler.CreateDataUploadAuditEntry)
			admin.GET("/audits/data-uploads", h.DataUploadAuditHandler.ListDataUploadAuditEntries)

			// Prize structure routes
			admin.GET("/prize-structures", h.AdminHandlers.ListPrizeStructures)
			admin.POST("/prize-structures", h.AdminHandlers.CreatePrizeStructure)
			admin.GET("/prize-structures/:id", h.AdminHandlers.GetPrizeStructure)
			admin.PUT("/prize-structures/:id", h.AdminHandlers.UpdatePrizeStructure)
			admin.DELETE("/prize-structures/:id", h.AdminHandlers.DeletePrizeStructure)

			// Draw routes
			admin.GET("/draws/eligibility", h.AdminHandlers.GetDrawEligibilityStats)
			admin.POST("/draws/execute", h.AdminHandlers.ExecuteDraw)
			admin.GET("/draws", h.AdminHandlers.ListDraws)
			admin.GET("/draws/:id", h.AdminHandlers.GetDraw)

			// Winner routes
			admin.GET("/winners", h.AdminHandlers.ListWinners)
			admin.GET("/winners/export", h.AdminHandlers.ExportWinners)
			admin.POST("/winners/:id/claim", h.AdminHandlers.ClaimPrize)
			admin.POST("/winners/:id/invoke-runner-up", h.AdminHandlers.InvokeRunnerUp)

			// Audit log routes
			admin.GET("/audit-logs", h.AdminHandlers.ListAuditLogs)
			admin.GET("/audit-logs/types", h.AdminHandlers.GetAuditLogTypes)
		}
	}

	return router
}

// healthCheck handler for the /health endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
