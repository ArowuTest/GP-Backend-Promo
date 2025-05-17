package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Configuration holds all application configuration
type Configuration struct {
	// Server settings
	ServerPort string
	ServerHost string
	
	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	
	// JWT settings
	JWTSecret  string
	JWTExpiry  time.Duration
	
	// Environment
	Environment string
	
	// PostHog settings
	PostHogAPIKey string
	PostHogHost   string
	
	// SMS Gateway settings
	SMSGatewayURL      string
	SMSGatewayUsername string
	SMSGatewayPassword string
	
	// MTN API settings
	MTNAPIBaseURL string
	MTNAPIKey     string
}

// Config is the global configuration instance
var Config Configuration

// DB is the global database connection
var DB *gorm.DB

// Initialize loads configuration from environment variables
func Initialize() error {
	// Load .env file if it exists
	_ = godotenv.Load()
	
	// Server settings
	Config.ServerPort = getEnv("SERVER_PORT", "8080")
	Config.ServerHost = getEnv("SERVER_HOST", "0.0.0.0")
	
	// Database settings
	Config.DBHost = getEnv("DB_HOST", "localhost")
	Config.DBPort = getEnv("DB_PORT", "5432")
	Config.DBUser = getEnv("DB_USER", "postgres")
	Config.DBPassword = getEnv("DB_PASSWORD", "postgres")
	Config.DBName = getEnv("DB_NAME", "mynumba_don_win")
	Config.DBSSLMode = getEnv("DB_SSLMODE", "disable")
	
	// JWT settings
	Config.JWTSecret = getEnv("JWT_SECRET", "mynumba_don_win_secret_key")
	jwtExpiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	Config.JWTExpiry = time.Duration(jwtExpiryHours) * time.Hour
	
	// Environment
	Config.Environment = getEnv("ENVIRONMENT", "development")
	
	// PostHog settings
	Config.PostHogAPIKey = getEnv("POSTHOG_API_KEY", "")
	Config.PostHogHost = getEnv("POSTHOG_HOST", "https://app.posthog.com")
	
	// SMS Gateway settings
	Config.SMSGatewayURL = getEnv("SMS_GATEWAY_URL", "")
	Config.SMSGatewayUsername = getEnv("SMS_GATEWAY_USERNAME", "")
	Config.SMSGatewayPassword = getEnv("SMS_GATEWAY_PASSWORD", "")
	
	// MTN API settings
	Config.MTNAPIBaseURL = getEnv("MTN_API_BASE_URL", "")
	Config.MTNAPIKey = getEnv("MTN_API_KEY", "")
	
	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
