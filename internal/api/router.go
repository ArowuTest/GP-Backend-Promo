package api

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	adminhandlers "github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
) 

// SetupRouter initializes and configures the Gin router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS Middleware Configuration
	config := cors.DefaultConfig()
	// Allow specific origins, methods, and headers
	config.AllowOrigins = []string{"https://gp-admin-promo.vercel.app", "http://localhost:3000"} // Add your Vercel frontend and local dev URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config) )

	// Simple health check route (optional, but good for Render health checks)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Backend Healthy"}) 
	})

	// API v1 Group
	apiV1 := router.Group("/api/v1")
	{
		// Authentication routes
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/login", adminhandlers.LoginAdminUser) // Assuming LoginAdminUser handles the logic
			// If you have a separate registration for admin, add it here
			// authRoutes.POST("/register", adminhandlers.RegisterAdminUser)
		}

		// Admin routes - protected by auth middleware
		// We will add the JWT middleware here once it's ready
		adminProtectedRoutes := apiV1.Group("/admin")
		adminProtectedRoutes.Use(auth.JWTMiddleware()) // Apply JWT middleware
		{
			// User Management (SuperAdmin only)
			userManagementRoutes := adminProtectedRoutes.Group("/users")
			userManagementRoutes.Use(auth.RoleAuthMiddleware(models.SuperAdminRole))
			{
				userManagementRoutes.POST("/", adminhandlers.CreateAdminUser)
				userManagementRoutes.GET("/", adminhandlers.ListAdminUsers)
				userManagementRoutes.GET("/:id", adminhandlers.GetAdminUser)
				userManagementRoutes.PUT("/:id", adminhandlers.UpdateAdminUser)
				userManagementRoutes.DELETE("/:id", adminhandlers.DeleteAdminUser)
				userManagementRoutes.PUT("/:id/status", adminhandlers.UpdateAdminUserStatus)
			}

			// Prize Structure Management (SuperAdmin or DrawAdmin)
			prizeRoutes := adminProtectedRoutes.Group("/prize-structures")
			prizeRoutes.Use(auth.RoleAuthMiddleware(models.SuperAdminRole, models.DrawAdminRole))
			{
				prizeRoutes.POST("/", adminhandlers.CreatePrizeStructure)
				prizeRoutes.GET("/", adminhandlers.ListPrizeStructures)
				prizeRoutes.GET("/:id", adminhandlers.GetPrizeStructure)
				prizeRoutes.PUT("/:id", adminhandlers.UpdatePrizeStructure)
				prizeRoutes.DELETE("/:id", adminhandlers.DeletePrizeStructure)
				prizeRoutes.PUT("/:id/activate", adminhandlers.ActivatePrizeStructure)
			}

			// Draw Management (DrawAdmin or SuperAdmin)
			drawRoutes := adminProtectedRoutes.Group("/draws")
			drawRoutes.Use(auth.RoleAuthMiddleware(models.DrawAdminRole, models.SuperAdminRole))
			{
				drawRoutes.POST("/execute", adminhandlers.ExecuteDraw)
				drawRoutes.GET("/", adminhandlers.ListDraws)
				drawRoutes.GET("/:id", adminhandlers.GetDrawDetails)
				drawRoutes.POST("/:id/rerun", adminhandlers.RerunDraw) // Placeholder
			}

			// Winner Management & Reporting (SuperAdmin, DrawAdmin, ViewOnlyAdmin)
			winnerRoutes := adminProtectedRoutes.Group("/winners")
			winnerRoutes.Use(auth.RoleAuthMiddleware(models.SuperAdminRole, models.DrawAdminRole, models.ViewOnlyAdminRole))
			{
				winnerRoutes.GET("/", adminhandlers.ListWinners)
				winnerRoutes.GET("/export/momo", adminhandlers.ExportWinnersForMoMo) // Placeholder
				winnerRoutes.PUT("/:id/payment-status", adminhandlers.UpdateWinnerPaymentStatus)
			}

			// Audit Logs (SuperAdmin)
			auditRoutes := adminProtectedRoutes.Group("/audit-logs")
			auditRoutes.Use(auth.RoleAuthMiddleware(models.SuperAdminRole))
			{
				auditRoutes.GET("/", adminhandlers.ListAuditLogs) // Placeholder
			}
		}
	}

	return router
}

