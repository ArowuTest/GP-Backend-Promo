package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

func main() {
	// Check command line arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: reset_password_tool <email> <new_password>")
		fmt.Println("Example: reset_password_tool admin@example.com NewSecurePassword123!")
		os.Exit(1)
	}
	
	email := os.Args[1]
	newPassword := os.Args[2]
	
	// Validate password
	if err := validatePassword(newPassword); err != nil {
		log.Fatalf("Password validation failed: %v", err)
	}
	
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	
	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	
	// Find user by email
	var userID string
	var username string
	err = db.QueryRow("SELECT id, username FROM users WHERE email = $1", email).Scan(&userID, &username)
	if err != nil {
		log.Fatalf("Failed to find user with email %s: %v", email, err)
	}
	
	// Generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	// Update user password
	_, err = db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", 
		string(hashedPassword), userID)
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}
	
	// Log audit entry
	auditID := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO audits (id, action, entity_type, entity_id, user_id, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())",
		auditID, "PASSWORD_RESET_EMERGENCY", "User", userID, userID, fmt.Sprintf("Emergency password reset for user %s", username),
	)
	if err != nil {
		// Just log the error but continue
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
	}
	
	fmt.Printf("Password successfully reset for user %s (%s)\n", username, email)
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*' || char == '(' || char == ')' || char == '-' || char == '_' || char == '+' || char == '=' || char == '{' || char == '}' || char == '[' || char == ']' || char == '|' || char == '\\' || char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' || char == '>' || char == ',' || char == '.' || char == '?' || char == '/':
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}
	
	return nil
}
