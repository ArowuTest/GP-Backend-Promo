package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/persistence/gorm"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/middleware"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/handler"

	// Application services
	auditApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	drawApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	participantApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	prizeApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	userApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Set up repositories
	auditRepo := gorm.NewGormAuditRepository(db.DB)
	drawRepo := gorm.NewGormDrawRepository(db.DB)
	participantRepo := gorm.NewGormParticipantRepository(db.DB)
	prizeRepo := gorm.NewGormPrizeRepository(db.DB)
	userRepo := gorm.NewGormUserRepository(db.DB)

	// Set up application services
	logAuditService := auditApp.NewLogAuditService(auditRepo)
	getAuditLogsService := auditApp.NewGetAuditLogsService(auditRepo)
	getDataUploadAuditsService := auditApp.NewGetDataUploadAuditsService(auditRepo)

	// Draw services
	executeDrawService := drawApp.NewDrawService(drawRepo, participantRepo, prizeRepo, logAuditService)
	getDrawByIDService := drawApp.NewGetDrawByIDService(drawRepo)
	listDrawsService := drawApp.NewListDrawsService(drawRepo)
	listWinnersService := drawApp.NewListWinnersService(drawRepo)
	getEligibilityStatsService := drawApp.NewGetEligibilityStatsService(drawRepo, participantRepo)
	invokeRunnerUpService := drawApp.NewInvokeRunnerUpService(drawRepo, logAuditService)
	updateWinnerPaymentStatusService := drawApp.NewUpdateWinnerPaymentStatusService(drawRepo)

	// Participant services
	uploadParticipantsService := participantApp.NewUploadParticipantsService(participantRepo, logAuditService)
	getParticipantStatsService := participantApp.NewGetParticipantStatsService(participantRepo)
	listParticipantsService := participantApp.NewListParticipantsService(participantRepo)
	listUploadAuditsService := participantApp.NewListUploadAuditsService(participantRepo)
	deleteUploadService := participantApp.NewDeleteUploadService(participantRepo)

	// Prize services
	createPrizeStructureService := prizeApp.NewCreatePrizeStructureService(prizeRepo, logAuditService)
	getPrizeStructureService := prizeApp.NewGetPrizeStructureService(prizeRepo)
	listPrizeStructuresService := prizeApp.NewListPrizeStructuresService(prizeRepo)
	updatePrizeStructureService := prizeApp.NewUpdatePrizeStructureService(prizeRepo, logAuditService)

	// User services
	authenticateUserService := userApp.NewAuthenticateUserService(userRepo, logAuditService)
	createUserService := userApp.NewCreateUserService(userRepo, logAuditService)
	updateUserService := userApp.NewUpdateUserService(userRepo, logAuditService)
	getUserService := userApp.NewGetUserService(userRepo)
	listUsersService := userApp.NewListUsersService(userRepo)
	
	// Password reset service
	resetPasswordService := userApp.NewResetPasswordService(userRepo, logAuditService)

	// Set up middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	corsMiddleware := middleware.Default()
	errorMiddleware := middleware.NewErrorMiddleware(true)

	// Set up gin engine
	ginEngine := gin.Default()

	// Set up handlers
	auditHandler := handler.NewAuditHandler(
		getAuditLogsService,
		getDataUploadAuditsService,
	)
	drawHandler := handler.NewDrawHandler(
		executeDrawService,
		getDrawByIDService,
		listDrawsService,
		listWinnersService,
		getEligibilityStatsService,
		invokeRunnerUpService,
		updateWinnerPaymentStatusService,
	)
	participantHandler := handler.NewParticipantHandler(
		listParticipantsService,
		getParticipantStatsService,
		listUploadAuditsService,
		uploadParticipantsService,
		deleteUploadService,
	)
	prizeHandler := handler.NewPrizeHandler(
		createPrizeStructureService,
		getPrizeStructureService,
		listPrizeStructuresService,
		updatePrizeStructureService,
	)
	userHandler := handler.NewUserHandler(
		authenticateUserService,
		createUserService,
		updateUserService,
		getUserService,
		listUsersService,
	)
	
	// Password reset handler
	resetPasswordHandler := handler.NewResetPasswordHandler(resetPasswordService)

	// Set up router
	router := api.NewRouter(
		ginEngine,
		authMiddleware,
		corsMiddleware,
		errorMiddleware,
		drawHandler,
		prizeHandler,
		participantHandler,
		auditHandler,
		userHandler,
		resetPasswordHandler,
	)

	// Setup routes
	router.Setup()

	// Set up server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      ginEngine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Run migrations
	if err := db.Migrate(
		&gorm.AuditLogModel{},
		&gorm.SystemAuditLogModel{},
		&gorm.DrawModel{},
		&gorm.WinnerModel{},
		&gorm.ParticipantModel{},
		&gorm.UploadAuditModel{},
		&gorm.PrizeStructureModel{},
		&gorm.PrizeTierModel{},
		&gorm.PrizeModel{},
		&gorm.UserModel{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Fatalf("Failed to close database connection: %v", err)
	}

	log.Println("Server exited properly")
}
