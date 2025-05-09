package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	myauth "github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	adminhandlers "github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Health check - can be enhanced to check DB status etc.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Backend Healthy from Gin Router"})
	})

	// API v1 Group
	apiV1 := router.Group("/api/v1")
	{
		// Authentication routes
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/login", myauth.LoginAdminUser)
		}

		// Admin routes - protected by auth middleware
		adminRoutes := apiV1.Group("/admin")
		adminRoutes.Use(myauth.JWTMiddleware()) // Apply JWT middleware
		{
			// User Management (SuperAdmin only)
			userManagementRoutes := adminRoutes.Group("/users")
			userManagementRoutes.Use(myauth.RoleAuthMiddleware(models.SuperAdminRole))
			{
				userManagementRoutes.POST("/", adminhandlers.CreateAdminUser)
				userManagementRoutes.GET("/", adminhandlers.ListAdminUsers)
				userManagementRoutes.GET("/:id", adminhandlers.GetAdminUser)
				userManagementRoutes.PUT("/:id", adminhandlers.UpdateAdminUser)
				userManagementRoutes.DELETE("/:id", adminhandlers.DeleteAdminUser)
				userManagementRoutes.PUT("/:id/status", adminhandlers.UpdateAdminUserStatus)
			}

			// Prize Structure Management (SuperAdmin or DrawAdmin)
			prizeRoutes := adminRoutes.Group("/prize-structures")
			prizeRoutes.Use(myauth.RoleAuthMiddleware(models.SuperAdminRole, models.DrawAdminRole))
			{
				prizeRoutes.POST("/", adminhandlers.CreatePrizeStructure)
				prizeRoutes.GET("/", adminhandlers.ListPrizeStructures)
				prizeRoutes.GET("/:id", adminhandlers.GetPrizeStructure)
				prizeRoutes.PUT("/:id", adminhandlers.UpdatePrizeStructure)
				prizeRoutes.DELETE("/:id", adminhandlers.DeletePrizeStructure)
				prizeRoutes.PUT("/:id/activate", adminhandlers.ActivatePrizeStructure)
			}

			// Draw Management (DrawAdmin)
			drawRoutes := adminRoutes.Group("/draws")
			drawRoutes.Use(myauth.RoleAuthMiddleware(models.DrawAdminRole, models.SuperAdminRole)) // Allow SuperAdmin too
			{
				drawRoutes.POST("/execute", adminhandlers.ExecuteDraw)
				drawRoutes.GET("/", adminhandlers.ListDraws)
				drawRoutes.GET("/:id", adminhandlers.GetDrawDetails)
				drawRoutes.POST("/:id/rerun", adminhandlers.RerunDraw) // Placeholder
			}

			// Winner Management & Reporting
			winnerRoutes := adminRoutes.Group("/winners")
			winnerRoutes.Use(myauth.RoleAuthMiddleware(models.SuperAdminRole, models.DrawAdminRole, models.ViewOnlyAdminRole))
			{
				winnerRoutes.GET("/", adminhandlers.ListWinners)
				winnerRoutes.GET("/export/momo", adminhandlers.ExportWinnersForMoMo) // Placeholder
				winnerRoutes.PUT("/:id/payment-status", adminhandlers.UpdateWinnerPaymentStatus)
			}

			// Audit Logs (SuperAdmin)
			auditRoutes := adminRoutes.Group("/audit-logs")
			auditRoutes.Use(myauth.RoleAuthMiddleware(models.SuperAdminRole))
			{
				auditRoutes.GET("/", adminhandlers.ListAuditLogs) // Placeholder
			}
		}
	}

	return router
}

