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
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/persistence/gorm"
)

// Container handles dependency injection for the application
type Container struct {
	// Database
	DB *gorm.DB
	
	// Repositories
	UserRepository        *gorm.GormUserRepository
	DrawRepository        *gorm.GormDrawRepository
	ParticipantRepository *gorm.GormParticipantRepository
	PrizeRepository       *gorm.GormPrizeRepository
	AuditRepository       *gorm.GormAuditRepository
	
	// Services
	AuthService           *user.AuthenticateUserService
	DrawService           *draw.DrawService
	ParticipantService    *participant.ParticipantService
	PrizeService          *prize.PrizeService
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
	c.UserRepository = gorm.NewGormUserRepository(c.DB)
	c.DrawRepository = gorm.NewGormDrawRepository(c.DB)
	c.ParticipantRepository = gorm.NewGormParticipantRepository(c.DB)
	c.PrizeRepository = gorm.NewGormPrizeRepository(c.DB)
	c.AuditRepository = gorm.NewGormAuditRepository(c.DB)
}

// Initialize services
func (c *Container) initServices() {
	c.AuditService = audit.NewAuditService(c.AuditRepository)
	c.AuthService = user.NewAuthenticateUserService(c.UserRepository, c.AuditService)
	c.DrawService = draw.NewDrawService(c.DrawRepository, c.ParticipantRepository, c.PrizeRepository, c.AuditService)
	c.ParticipantService = participant.NewParticipantService(c.ParticipantRepository, c.AuditService)
	c.PrizeService = prize.NewPrizeService(c.PrizeRepository, c.AuditService)
	c.ResetPasswordService = user.NewResetPasswordService(c.UserRepository, c.AuditService)
}

// Initialize middleware
func (c *Container) initMiddleware() {
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.AuthService)
	c.CORSMiddleware = middleware.NewCORSMiddleware()
	c.ErrorMiddleware = middleware.NewErrorMiddleware()
}

// Initialize handlers
func (c *Container) initHandlers() {
	c.DrawHandler = handler.NewDrawHandler(c.DrawService)
	c.PrizeHandler = handler.NewPrizeHandler(c.PrizeService)
	c.ParticipantHandler = handler.NewParticipantHandler(c.ParticipantService)
	c.AuditHandler = handler.NewAuditHandler(c.AuditService)
	c.UserHandler = handler.NewUserHandler(c.AuthService)
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
		c.ResetPasswordHandler,
	)
}

// Setup configures the application
func (c *Container) Setup() {
	c.Router.Setup()
}

// Run starts the HTTP server
func (c *Container) Run(addr string) error {
	return c.Router.Run(addr)
}
