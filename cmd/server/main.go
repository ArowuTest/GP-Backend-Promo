package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/infrastructure/persistence/gorm"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/handler"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/middleware"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize database connection
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	drawRepo := infrastructure.NewGormDrawRepository(db)
	participantRepo := infrastructure.NewGormParticipantRepository(db)
	prizeRepo := infrastructure.NewGormPrizeRepository(db)
	auditRepo := infrastructure.NewGormAuditRepository(db)
	userRepo := infrastructure.NewGormUserRepository(db)
	uploadAuditRepo := infrastructure.NewGormUploadAuditRepository(db)

	// Initialize use cases
	executeDraw := application.NewExecuteDrawUseCase(drawRepo, prizeRepo, participantRepo)
	getDrawByID := application.NewGetDrawByIDUseCase(drawRepo)
	listDraws := application.NewListDrawsUseCase(drawRepo)
	getEligibilityStats := application.NewGetEligibilityStatsUseCase(drawRepo, participantRepo)
	invokeRunnerUp := application.NewInvokeRunnerUpUseCase(drawRepo)
	
	uploadParticipants := application.NewUploadParticipantsUseCase(participantRepo, uploadAuditRepo)
	getParticipantStats := application.NewGetParticipantStatsUseCase(participantRepo)
	listParticipants := application.NewListParticipantsUseCase(participantRepo)
	deleteUpload := application.NewDeleteUploadUseCase(uploadAuditRepo, participantRepo, auditRepo)
	
	createPrizeStructure := application.NewCreatePrizeStructureUseCase(prizeRepo)
	updatePrizeStructure := application.NewUpdatePrizeStructureUseCase(prizeRepo)
	getPrizeStructure := application.NewGetPrizeStructureUseCase(prizeRepo)
	listPrizeStructures := application.NewListPrizeStructuresUseCase(prizeRepo)
	deletePrizeStructure := application.NewDeletePrizeStructureUseCase(prizeRepo)
	
	logAudit := application.NewLogAuditUseCase(auditRepo)
	getAuditLogs := application.NewGetAuditLogsUseCase(auditRepo)
	getDataUploadAudits := application.NewGetDataUploadAuditsUseCase(uploadAuditRepo)
	
	authenticateUser := application.NewAuthenticateUserUseCase(userRepo)
	createUser := application.NewCreateUserUseCase(userRepo)
	updateUser := application.NewUpdateUserUseCase(userRepo)
	getUserByID := application.NewGetUserByIDUseCase(userRepo)
	listUsers := application.NewListUsersUseCase(userRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	corsMiddleware := middleware.Default()
	errorMiddleware := middleware.NewErrorMiddleware(cfg.Environment == "development")

	// Initialize handlers
	drawHandler := handler.NewDrawHandler(
		executeDraw,
		getDrawByID,
		listDraws,
		getEligibilityStats,
		invokeRunnerUp,
	)
	
	participantHandler := handler.NewParticipantHandler(
		uploadParticipants,
		listParticipants,
		getParticipantStats,
		deleteUpload,
	)
	
	prizeHandler := handler.NewPrizeHandler(
		createPrizeStructure,
		updatePrizeStructure,
		getPrizeStructure,
		listPrizeStructures,
		deletePrizeStructure,
	)
	
	auditHandler := handler.NewAuditHandler(
		logAudit,
		getAuditLogs,
		getDataUploadAudits,
	)
	
	userHandler := handler.NewUserHandler(
		authenticateUser,
		createUser,
		updateUser,
		getUserByID,
		listUsers,
	)

	// Initialize router
	engine := gin.Default()
	router := api.NewRouter(
		engine,
		authMiddleware,
		corsMiddleware,
		errorMiddleware,
		drawHandler,
		prizeHandler,
		participantHandler,
		auditHandler,
		userHandler,
	)

	// Setup routes
	router.Setup()

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	// Auto migrate models
	// This is for development only, production should use proper migrations
	if cfg.Environment == "development" {
		// Add auto-migration for all models
		// db.AutoMigrate(&models.Draw{}, &models.Winner{}, ...)
	}
	
	return db, nil
}
