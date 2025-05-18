package api

import (
	"github.com/gin-gonic/gin"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/handler"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/middleware"
)

// Router handles API routing
type Router struct {
	engine           *gin.Engine
	authMiddleware   *middleware.AuthMiddleware
	corsMiddleware   *middleware.CORSMiddleware
	errorMiddleware  *middleware.ErrorMiddleware
	drawHandler      *handler.DrawHandler
	prizeHandler     *handler.PrizeHandler
	participantHandler *handler.ParticipantHandler
	auditHandler     *handler.AuditHandler
	userHandler      *handler.UserHandler
}

// NewRouter creates a new Router
func NewRouter(
	engine *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	errorMiddleware *middleware.ErrorMiddleware,
	drawHandler *handler.DrawHandler,
	prizeHandler *handler.PrizeHandler,
	participantHandler *handler.ParticipantHandler,
	auditHandler *handler.AuditHandler,
	userHandler *handler.UserHandler,
) *Router {
	return &Router{
		engine:           engine,
		authMiddleware:   authMiddleware,
		corsMiddleware:   corsMiddleware,
		errorMiddleware:  errorMiddleware,
		drawHandler:      drawHandler,
		prizeHandler:     prizeHandler,
		participantHandler: participantHandler,
		auditHandler:     auditHandler,
		userHandler:      userHandler,
	}
}

// Setup configures all routes
func (r *Router) Setup() {
	// Apply global middleware
	r.engine.Use(r.corsMiddleware.Handle())
	r.engine.Use(r.errorMiddleware.Recovery())
	r.engine.Use(r.errorMiddleware.Handle())
	
	// Health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	
	// API v1 group
	api := r.engine.Group("/api/v1")
	
	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/login", r.userHandler.Login)
	}
	
	// Admin routes (require authentication)
	admin := api.Group("/admin")
	admin.Use(r.authMiddleware.Authenticate())
	{
		// Draw routes
		draws := admin.Group("/draws")
		{
			draws.GET("/eligibility-stats", r.drawHandler.GetEligibilityStats)
			draws.POST("/execute", r.authMiddleware.RequireRole("super_admin"), r.drawHandler.ExecuteDraw)
			draws.POST("/invoke-runner-up", r.authMiddleware.RequireRole("super_admin", "admin"), r.drawHandler.InvokeRunnerUp)
			draws.GET("", r.drawHandler.ListDraws)
			draws.GET("/:id", r.drawHandler.GetDrawByID)
		}
		
		// Winner routes
		winners := admin.Group("/winners")
		{
			winners.GET("", r.drawHandler.ListWinners)
			winners.PUT("/:id/payment-status", r.authMiddleware.RequireRole("super_admin", "admin"), r.drawHandler.UpdateWinnerPaymentStatus)
		}
		
		// Prize structure routes
		prizeStructures := admin.Group("/prize-structures")
		{
			prizeStructures.GET("", r.prizeHandler.ListPrizeStructures)
			prizeStructures.POST("", r.authMiddleware.RequireRole("super_admin", "admin"), r.prizeHandler.CreatePrizeStructure)
			prizeStructures.GET("/:id", r.prizeHandler.GetPrizeStructure)
			prizeStructures.PUT("/:id", r.authMiddleware.RequireRole("super_admin", "admin"), r.prizeHandler.UpdatePrizeStructure)
			prizeStructures.DELETE("/:id", r.authMiddleware.RequireRole("super_admin"), r.prizeHandler.DeletePrizeStructure)
		}
		
		// Participant routes
		participants := admin.Group("/participants")
		{
			participants.POST("/upload", r.authMiddleware.RequireRole("super_admin", "admin", "senior_user"), r.participantHandler.UploadParticipants)
			participants.GET("/stats", r.participantHandler.GetParticipantStats)
			participants.GET("/uploads", r.participantHandler.ListUploadAudits)
			participants.GET("", r.participantHandler.ListParticipants)
			participants.DELETE("/uploads/:id", r.authMiddleware.RequireRole("super_admin", "admin"), r.participantHandler.DeleteUpload)
		}
		
		// Report routes
		reports := admin.Group("/reports")
		{
			reports.GET("/data-uploads", r.auditHandler.GetDataUploadAudits)
		}
		
		// User routes
		users := admin.Group("/users")
		{
			users.GET("", r.authMiddleware.RequireRole("super_admin"), r.userHandler.ListUsers)
			users.POST("", r.authMiddleware.RequireRole("super_admin"), r.userHandler.CreateUser)
			users.GET("/:id", r.authMiddleware.RequireRole("super_admin"), r.userHandler.GetUserByID)
			users.PUT("/:id", r.authMiddleware.RequireRole("super_admin"), r.userHandler.UpdateUser)
		}
	}
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
