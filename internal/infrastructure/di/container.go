package di

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/handler"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/middleware"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api"
	pgorm "github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/persistence/gorm"
)

// Container handles dependency injection for the application
type Container struct {
	// Database
	DB *gorm.DB
	
	// Repositories
	UserRepository        *pgorm.GormUserRepository
	DrawRepository        *pgorm.GormDrawRepository
	ParticipantRepository *pgorm.GormParticipantRepository
	PrizeRepository       *pgorm.GormPrizeRepository
	AuditRepository       *pgorm.GormAuditRepository
	
	// Services
	AuthService           *user.AuthenticateUserService
	DrawService           *draw.ExecuteDrawService
	ParticipantService    *participant.UploadParticipantsService
	PrizeService          *prize.CreatePrizeStructureService
	AuditService          *audit.AuditService
	ResetPasswordService  *user.ResetPasswordService
	
	// Middleware
	AuthMiddleware        *middleware.AuthMiddleware
	CORSMiddleware        *middleware.CORSMiddleware
	ErrorMiddleware       *middleware.ErrorMiddleware
	
	// Handlers
	DrawHandler           *handler.DrawHandler
	PrizeHandler          *handler.PrizeHandler
	ParticipantHandler    *handler.ParticipantHandler
	AuditHandler          *handler.AuditHandler
	UserHandler           *handler.UserHandler
	ResetPasswordHandler  *handler.ResetPasswordHandler
	
	// Router
	Router                *api.Router
	Engine                *gin.Engine
}

// NewContainer creates a new dependency injection container
func NewContainer(db *gorm.DB) *Container {
	container := &Container{
		DB:     db,
		Engine: gin.Default(),
	}
	
	// Initialize repositories
	container.initRepositories()
	
	// Initialize services
	container.initServices()
	
	// Initialize middleware
	container.initMiddleware()
	
	// Initialize handlers
	container.initHandlers()
	
	// Initialize router
	container.initRouter()
	
	return container
}

// Initialize repositories
func (c *Container) initRepositories() {
	c.UserRepository = pgorm.NewGormUserRepository(c.DB)
	c.DrawRepository = pgorm.NewGormDrawRepository(c.DB)
	c.ParticipantRepository = pgorm.NewGormParticipantRepository(c.DB)
	c.PrizeRepository = pgorm.NewGormPrizeRepository(c.DB)
	c.AuditRepository = pgorm.NewGormAuditRepository(c.DB)
}

// Initialize services
func (c *Container) initServices() {
	logAuditService := audit.NewLogAuditService(c.AuditRepository)
	c.AuditService = audit.NewAuditService(logAuditService)
	
	// Create audit-related services
	getAuditLogsService := audit.NewGetAuditLogsService(c.AuditRepository)
	getDataUploadAuditsService := audit.NewGetDataUploadAuditsService(c.AuditRepository)
	
	c.AuthService = user.NewAuthenticateUserService(c.UserRepository, c.AuditService)
	c.DrawService = draw.NewDrawService(c.DrawRepository, c.ParticipantRepository, c.PrizeRepository, c.AuditService)
	c.ParticipantService = participant.NewUploadParticipantsService(c.ParticipantRepository, c.AuditService)
	c.PrizeService = prize.NewCreatePrizeStructureService(c.PrizeRepository, c.AuditService)
	c.ResetPasswordService = user.NewResetPasswordService(c.UserRepository, c.AuditService)
}

// Initialize middleware
func (c *Container) initMiddleware() {
	c.AuthMiddleware = middleware.NewAuthMiddleware("mynumba-donwin-jwt-secret-key-2025") // Production JWT secret
	c.CORSMiddleware = middleware.NewCORSMiddleware(
		[]string{"https://gp-admin-promo.vercel.app"},
		[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		[]string{"Content-Type", "Authorization", "Accept"},
		[]string{},
		true)
	c.ErrorMiddleware = middleware.NewErrorMiddleware(false) // Set to true for debug mode
}

// Initialize handlers
func (c *Container) initHandlers() {
	// Note: These are simplified handler initializations that need to be expanded
	// with all required service dependencies in a production environment
	c.DrawHandler = handler.NewDrawHandler(
		c.DrawService,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
		
	c.PrizeHandler = handler.NewPrizeHandler(
		c.PrizeService,
		nil,
		nil,
		nil)
		
	c.ParticipantHandler = handler.NewParticipantHandler(
		nil,
		nil,
		nil,
		c.ParticipantService,
		nil)
		
	// Create audit-related services for handler
	getAuditLogsService := audit.NewGetAuditLogsService(c.AuditRepository)
	getDataUploadAuditsService := audit.NewGetDataUploadAuditsService(c.AuditRepository)
	
	c.AuditHandler = handler.NewAuditHandler(
		getAuditLogsService,
		getDataUploadAuditsService)
		
	c.UserHandler = handler.NewUserHandler(
		c.AuthService,
		nil,
		nil,
		nil,
		nil)
		
	c.ResetPasswordHandler = handler.NewResetPasswordHandler(c.ResetPasswordService)
}

// Initialize router
func (c *Container) initRouter() {
	c.Router = api.NewRouter(
		c.Engine,
		c.AuthMiddleware,
		c.CORSMiddleware,
		c.ErrorMiddleware,
		c.DrawHandler,
		c.PrizeHandler,
		c.ParticipantHandler,
		c.AuditHandler,
		c.UserHandler,
		c.ResetPasswordHandler)
}

// Setup configures the application
func (c *Container) Setup() {
	c.Router.Setup()
}

// Run starts the HTTP server
func (c *Container) Run(addr string) error {
	return c.Router.Run(addr)
}
