package config

import (
	"fmt"
	"os"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	err := godotenv.Load(".env.development") // In production, Render will set these
	if err != nil {
		fmt.Println("Error loading .env file, relying on OS environment variables")
	}
}

// ConnectDB connects to the PostgreSQL database
func ConnectDB() {
	LoadEnv()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback DSN for local development if DATABASE_URL is not set
		dsn = "host=localhost user=postgres password=postgres dbname=mynumba_dev port=5432 sslmode=disable TimeZone=UTC"
		fmt.Println("DATABASE_URL not set, using default local DSN")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Use logger.Info for dev, logger.Silent for prod
	})

	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	fmt.Println("Database connected successfully")

	// Auto-migrate the schema
	autoMigrate()
}

func autoMigrate() {
	// Enable UUID extension if not enabled. This should ideally be done by a superuser once.
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		fmt.Println("Warning: Failed to create uuid-ossp extension, it might already exist or require superuser privileges: ", err)
	}

	// Auto-migrate all models
	err := DB.AutoMigrate(
		&models.AdminUser{},
		&models.PrizeStructure{},
		&models.Prize{},
		&models.Draw{},
		&models.DrawWinner{},
		&models.Participant{}, // Added Participant model
		&models.DataUploadAudit{}, // Added DataUploadAudit model
		// Add any other models here that need to be migrated
	)

	if err != nil {
		panic("Failed to auto-migrate database schema: " + err.Error())
	}
	fmt.Println("Database migration completed successfully")
}

