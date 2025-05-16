package api

import (
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers"      // General handlers
	admin_handlers "github.com/ArowuTest/GP-Backend-Promo/internal/handlers/admin" // Alias for admin sub-package handlers
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/ArowuTest/GP-Backend-Promo/internal/services"      // Added for DrawDataService
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Disable automatic trailing slash redirects to prevent CORS preflight issues
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// CORS Middleware Configuration - Enhanced for production
	// Use a more permissive CORS configuration to ensure all requests work
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins temporarily for debugging
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Custom middleware to handle OPTIONS requests for all paths
	// This ensures CORS preflight requests are properly handled
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*") // Allow all origins temporarily
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "43200") // 12 hours in seconds
			c.Status(http.StatusOK)
			c.Abort()
			return
		}
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Backend Healthy"})
	})

	// Instantiate services and handlers
	// For now, we use MockDrawDataService. This can be replaced with a real one later.
	drawService := &services.MockDrawDataService{}
	drawHandler := admin_handlers.NewDrawHandler(drawService)

	// API v1 Group
	apiV1 := router.Group("/api/v1")

	// Authentication routes (public)
	authRoutes := apiV1.Group("/auth")
	authRoutes.POST("/login", admin_handlers.Login)

	// Admin routes - protected by JWT middleware
	adminProtectedRoutes := apiV1.Group("/admin")
	adminProtectedRoutes.Use(auth.JWTMiddleware())

	// User Management (SuperAdmin only)
	userManagement := adminProtectedRoutes.Group("/users")
	userManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin))
	userManagement.POST("/", admin_handlers.CreateAdminUser)
	userManagement.GET("/", admin_handlers.ListAdminUsers)
	userManagement.GET("/:id", admin_handlers.GetAdminUser)
	userManagement.PUT("/:id", admin_handlers.UpdateAdminUser)
	userManagement.DELETE("/:id", admin_handlers.DeleteAdminUser)

	// Prize Structure Management (SuperAdmin, Admin)
	prizeManagement := adminProtectedRoutes.Group("/prize-structures")
	prizeManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin))
	
	// Create prize structure - both with and without trailing slash
	prizeManagement.POST("/", admin_handlers.CreatePrizeStructure)
	prizeManagement.POST("", admin_handlers.CreatePrizeStructure)
	
	// List prize structures - both with and without trailing slash
	prizeManagement.GET("/", admin_handlers.ListPrizeStructures)
	prizeManagement.GET("", admin_handlers.ListPrizeStructures)
	
	// Get single prize structure - both with and without trailing slash
	prizeManagement.GET("/:id", admin_handlers.GetPrizeStructure)
	prizeManagement.GET("/:id/", admin_handlers.GetPrizeStructure)
	
	// Update prize structure - both with and without trailing slash
	prizeManagement.PUT("/:id", admin_handlers.UpdatePrizeStructure)
	prizeManagement.PUT("/:id/", admin_handlers.UpdatePrizeStructure)
	
	// Delete prize structure - both with and without trailing slash
	prizeManagement.DELETE("/:id", admin_handlers.DeletePrizeStructure)
	prizeManagement.DELETE("/:id/", admin_handlers.DeletePrizeStructure)

	// Draw Management
	drawManagement := adminProtectedRoutes.Group("/draws")
	// Use methods from the instantiated drawHandler
	drawManagement.POST("/execute", auth.RoleAuthMiddleware(models.RoleSuperAdmin), drawHandler.ExecuteDraw)
	drawManagement.POST("/invoke-runner-up", auth.RoleAuthMiddleware(models.RoleSuperAdmin), drawHandler.InvokeRunnerUp) // Added InvokeRunnerUp route
	drawManagement.GET("/", auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser), drawHandler.ListDraws)
	drawManagement.GET("/:draw_id", auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser), drawHandler.GetDrawDetails) // Changed :id to :draw_id to match handler

	// Participant Data Management (SuperAdmin, Admin)
	participantManagement := adminProtectedRoutes.Group("/participants")
	participantManagement.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin))
	participantManagement.POST("/upload", admin_handlers.HandleParticipantUpload)
	// Removed the GetParticipantStats endpoint that doesn't exist in the backend

	// Reporting
	reports := adminProtectedRoutes.Group("/reports")

	// Data Upload Audit Reporting (SuperAdmin, Admin, AllReportUser)
	dataUploadAudits := reports.Group("/data-uploads")
	dataUploadAudits.Use(auth.RoleAuthMiddleware(models.RoleSuperAdmin, models.RoleAdmin, models.RoleAllReportUser))
	dataUploadAudits.GET("/", handlers.ListDataUploadAuditEntries)

	return router
}
