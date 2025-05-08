package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mynumba_don_win_draw_system_backend_internal_auth "mynumba-don-win-draw-system/backend/internal/auth"
	mynumba_don_win_draw_system_backend_internal_handlers_admin "mynumba-don-win-draw-system/backend/internal/handlers/admin"
	mynumba_don_win_draw_system_backend_internal_models "mynumba-don-win-draw-system/backend/internal/models"
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
			authRoutes.POST("/login", mynumba_don_win_draw_system_backend_internal_auth.LoginAdminUser)
		}

		// Admin routes - protected by auth middleware
		adminRoutes := apiV1.Group("/admin")
		adminRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.JWTMiddleware()) // Apply JWT middleware
		{
			// User Management (SuperAdmin only)
			userManagementRoutes := adminRoutes.Group("/users")
			userManagementRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.RoleAuthMiddleware(mynumba_don_win_draw_system_backend_internal_models.SuperAdminRole))
			{
				userManagementRoutes.POST("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.CreateAdminUser)
				userManagementRoutes.GET("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.ListAdminUsers)
				userManagementRoutes.GET("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.GetAdminUser)
				userManagementRoutes.PUT("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.UpdateAdminUser)
				userManagementRoutes.DELETE("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.DeleteAdminUser)
				userManagementRoutes.PUT("/:id/status", mynumba_don_win_draw_system_backend_internal_handlers_admin.UpdateAdminUserStatus)
			}

			// Prize Structure Management (SuperAdmin or DrawAdmin)
			prizeRoutes := adminRoutes.Group("/prize-structures")
			prizeRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.RoleAuthMiddleware(mynumba_don_win_draw_system_backend_internal_models.SuperAdminRole, mynumba_don_win_draw_system_backend_internal_models.DrawAdminRole))
			{
				prizeRoutes.POST("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.CreatePrizeStructure)
				prizeRoutes.GET("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.ListPrizeStructures)
				prizeRoutes.GET("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.GetPrizeStructure)
				prizeRoutes.PUT("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.UpdatePrizeStructure)
				prizeRoutes.DELETE("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.DeletePrizeStructure)
				prizeRoutes.PUT("/:id/activate", mynumba_don_win_draw_system_backend_internal_handlers_admin.ActivatePrizeStructure)
			}

			// Draw Management (DrawAdmin)
			drawRoutes := adminRoutes.Group("/draws")
			drawRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.RoleAuthMiddleware(mynumba_don_win_draw_system_backend_internal_models.DrawAdminRole, mynumba_don_win_draw_system_backend_internal_models.SuperAdminRole)) // Allow SuperAdmin too
			{
				drawRoutes.POST("/execute", mynumba_don_win_draw_system_backend_internal_handlers_admin.ExecuteDraw)
				drawRoutes.GET("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.ListDraws)
				drawRoutes.GET("/:id", mynumba_don_win_draw_system_backend_internal_handlers_admin.GetDrawDetails)
				drawRoutes.POST("/:id/rerun", mynumba_don_win_draw_system_backend_internal_handlers_admin.RerunDraw) // Placeholder
			}

			// Winner Management & Reporting
			winnerRoutes := adminRoutes.Group("/winners")
			winnerRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.RoleAuthMiddleware(mynumba_don_win_draw_system_backend_internal_models.SuperAdminRole, mynumba_don_win_draw_system_backend_internal_models.DrawAdminRole, mynumba_don_win_draw_system_backend_internal_models.ViewOnlyAdminRole))
			{
				winnerRoutes.GET("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.ListWinners)
				winnerRoutes.GET("/export/momo", mynumba_don_win_draw_system_backend_internal_handlers_admin.ExportWinnersForMoMo) // Placeholder
				winnerRoutes.PUT("/:id/payment-status", mynumba_don_win_draw_system_backend_internal_handlers_admin.UpdateWinnerPaymentStatus)
			}

			// Audit Logs (SuperAdmin)
			auditRoutes := adminRoutes.Group("/audit-logs")
			auditRoutes.Use(mynumba_don_win_draw_system_backend_internal_auth.RoleAuthMiddleware(mynumba_don_win_draw_system_backend_internal_models.SuperAdminRole))
			{
				auditRoutes.GET("/", mynumba_don_win_draw_system_backend_internal_handlers_admin.ListAuditLogs) // Placeholder
			}
		}
	}

	return router
}

