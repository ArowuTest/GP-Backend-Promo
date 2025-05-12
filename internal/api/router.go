package api

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers" // General handlers
	admin_handlers "github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin" // Alias for admin sub-package handlers
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS Middleware Configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://gp-admin-promo.vercel.app"}, // Allow localhost for dev and Vercel for prod
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Backend Healthy"})
	})

	// API v1 Group
	apiV1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		authRoutes := apiV1.Group("/auth")
		{
			// Use Login from admin_handlers for consistency with other admin user functions
			authRoutes.POST("/login", admin_handlers.Login) 
		}

		// Admin routes - protected by JWT middleware
		adminProtectedRoutes := apiV1.Group("/admin")
		adminProtectedRoutes.Use(auth.JWTMiddleware())
		{
			// User Management (SuperAdmin only)
			userManagement := adminProtectedRoutes.Group("/users")
			userManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin))
			{
				userManagement.POST("/", admin_handlers.CreateAdminUser) // Corrected name
				userManagement.GET("/", admin_handlers.ListAdminUsers)    // Corrected name
				userManagement.GET("/:id", admin_handlers.GetAdminUser)     // Corrected name
				userManagement.PUT("/:id", admin_handlers.UpdateAdminUser)   // Corrected name
				userManagement.DELETE("/:id", admin_handlers.DeleteAdminUser) // Corrected name
			}

			// Prize Structure Management (SuperAdmin, Admin)
			prizeManagement := adminProtectedRoutes.Group("/prize-structures")
			prizeManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin))
			{
				prizeManagement.POST("/", admin_handlers.CreatePrizeStructure)
				prizeManagement.GET("/", admin_handlers.ListPrizeStructures)
				prizeManagement.GET("/:id", admin_handlers.GetPrizeStructure)
				prizeManagement.PUT("/:id", admin_handlers.UpdatePrizeStructure)
				prizeManagement.DELETE("/:id", admin_handlers.DeletePrizeStructure)
			}

			// Draw Management
			drawManagement := adminProtectedRoutes.Group("/draws")
			{
				drawManagement.POST("/execute", auth.RoleAuthMiddleware(models.RoleSuperAdmin), admin_handlers.ExecuteDraw)
				drawManagement.GET("/", auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser), admin_handlers.ListDraws)
				drawManagement.GET("/:id", auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser), admin_handlers.GetDrawDetails)
			}

			// Participant Data Management (SuperAdmin, Admin)
			participantManagement := adminProtectedRoutes.Group("/participants")
			participantManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin))
			{
				participantManagement.POST("/upload", admin_handlers.HandleParticipantUpload)
			}

			// Reporting
			reports := adminProtectedRoutes.Group("/reports")
			{
				// Winner Reporting (All roles that need winner reports)
				// winnerReports := reports.Group("/winners")
				// winnerReports.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser))
				// {
				// 	// winnerReports.GET("/", admin_handlers.ListWinners) // Assuming ListWinners is in admin_handlers
				// }

				// Data Upload Audit Reporting (SuperAdmin, Admin, AllReportUser)
				dataUploadAudits := reports.Group("/data-uploads")
				dataUploadAudits.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleAllReportUser))
				{
					// ListDataUploadAuditEntries is in package handlers, not admin_handlers
					dataUploadAudits.GET("/", handlers.ListDataUploadAuditEntries) 
				}
			}
		}
	}

	return router
}

