package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	
	// JWT configuration
	JWTSecret        string
	JWTExpirationHours int
	
	// API configuration
	APIPort     string
	Environment string
	
	// External services
	PostHogAPIKey string
	PostHogHost   string
	SMSGatewayURL string
	SMSGatewayAPIKey string
	MTNApiURL    string
	MTNApiKey    string
}

// NewConfig creates a new Config with values from environment variables
func NewConfig() *Config {
	jwtExpirationHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	
	return &Config{
		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "gp_backend_promo"),
		
		// JWT configuration
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-here"),
		JWTExpirationHours: jwtExpirationHours,
		
		// API configuration
		APIPort:     getEnv("API_PORT", "8080"),
		Environment: getEnv("API_ENV", "development"),
		
		// External services
		PostHogAPIKey:    getEnv("POSTHOG_API_KEY", ""),
		PostHogHost:      getEnv("POSTHOG_HOST", "https://app.posthog.com"),
		SMSGatewayURL:    getEnv("SMS_GATEWAY_URL", ""),
		SMSGatewayAPIKey: getEnv("SMS_GATEWAY_API_KEY", ""),
		MTNApiURL:        getEnv("MTN_API_URL", ""),
		MTNApiKey:        getEnv("MTN_API_KEY", ""),
	}
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
