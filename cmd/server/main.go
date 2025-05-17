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
)

// Services holds all service instances
type Services struct {
	AuditService    *services.AuditService
	DrawDataService services.DrawDataService
	// Add other services here
}

// Handlers holds all handler instances
type Handlers struct {
	AdminHandlers *admin.AdminHandlers
	// Add other handlers here
}

func main() {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Initialize database connection
	dbInstance, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize services with DB dependency injection
	services := setupServices(dbInstance.DB)

	// Initialize handlers with services
	handlers := setupHandlers(services)

	// Setup router with handlers
	router := setupRouter(handlers)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = config.Config.ServerPort
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupServices initializes all services with DB dependency injection
func setupServices(db *gorm.DB) *Services {
	return &Services{
		AuditService:    services.NewAuditService(db),
		DrawDataService: services.NewDatabaseDrawDataService(db),
		// Initialize other services with DB here
	}
}

// setupHandlers initializes all handlers with service dependencies
func setupHandlers(services *Services) *Handlers {
	return &Handlers{
		AdminHandlers: admin.NewAdminHandlers(
			services.AuditService,
			services.DrawDataService,
			// Pass other required services here
		),
		// Initialize other handlers here
	}
}

// setupRouter configures the HTTP router with all routes and middleware
func setupRouter(handlers *Handlers) *gin.Engine {
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
		api.GET("/health", handlers.HealthCheck)

		// Admin routes (protected)
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth())
		{
			// Prize structure routes
			admin.GET("/prize-structures", handlers.AdminHandlers.ListPrizeStructures)
			admin.POST("/prize-structures", handlers.AdminHandlers.CreatePrizeStructure)
			admin.GET("/prize-structures/:id", handlers.AdminHandlers.GetPrizeStructure)
			admin.PUT("/prize-structures/:id", handlers.AdminHandlers.UpdatePrizeStructure)
			admin.DELETE("/prize-structures/:id", handlers.AdminHandlers.DeletePrizeStructure)

			// Draw routes
			admin.GET("/draws/eligibility", handlers.AdminHandlers.GetDrawEligibilityStats)
			admin.POST("/draws/execute", handlers.AdminHandlers.ExecuteDraw)
			admin.GET("/draws", handlers.AdminHandlers.ListDraws)
			admin.GET("/draws/:id", handlers.AdminHandlers.GetDraw)

			// Winner routes
			admin.GET("/winners", handlers.AdminHandlers.ListWinners)
			admin.GET("/winners/export", handlers.AdminHandlers.ExportWinners)
			admin.POST("/winners/:id/claim", handlers.AdminHandlers.ClaimPrize)
			admin.POST("/winners/:id/invoke-runner-up", handlers.AdminHandlers.InvokeRunnerUp)

			// Audit log routes
			admin.GET("/audit-logs", handlers.AdminHandlers.ListAuditLogs)
			admin.GET("/audit-logs/types", handlers.AdminHandlers.GetAuditLogTypes)

			// User management routes
			admin.GET("/users", handlers.AdminHandlers.ListUsers)
			admin.POST("/users", handlers.AdminHandlers.CreateUser)
			admin.PUT("/users/:id", handlers.AdminHandlers.UpdateUser)
			admin.DELETE("/users/:id", handlers.AdminHandlers.DeleteUser)
		}

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.AdminHandlers.Login)
			auth.POST("/refresh", handlers.AdminHandlers.RefreshToken)
		}
	}

	return router
}

// HealthCheck handler for the /health endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
