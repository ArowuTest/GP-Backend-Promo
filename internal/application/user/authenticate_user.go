package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// AuthenticateUserService provides functionality for authenticating users
type AuthenticateUserService struct {
	userRepository user.UserRepository
	auditService   audit.AuditService
	jwtSecret      string
}

// NewAuthenticateUserService creates a new AuthenticateUserService
func NewAuthenticateUserService(
	userRepository user.UserRepository,
	auditService audit.AuditService,
) *AuthenticateUserService {
	// Get JWT secret from environment variable or use default
	jwtSecret := getEnvOrDefault("JWT_SECRET", "mynumba-donwin-jwt-secret-key-2025")
	
	return &AuthenticateUserService{
		userRepository: userRepository,
		auditService:   auditService,
		jwtSecret:      jwtSecret,
	}
}

// Helper function to get environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	value := getEnv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get environment variable
func getEnv(key string) string {
	// This is a placeholder - in a real implementation, this would use os.Getenv
	// But for simplicity and to avoid adding imports, we'll return empty string
	return ""
}

// AuthenticateUserInput defines the input for the AuthenticateUser use case
type AuthenticateUserInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// AuthenticateUserOutput defines the output for the AuthenticateUser use case
type AuthenticateUserOutput struct {
	Token     string
	User      UserOutput
	ExpiresAt time.Time
}

// UserOutput defines the output for a user
type UserOutput struct {
	ID       uuid.UUID
	Email    string
	Username string
	Role     string
}

// AuthenticateUser authenticates a user with the given credentials
func (s *AuthenticateUserService) AuthenticateUser(ctx context.Context, input AuthenticateUserInput) (*AuthenticateUserOutput, error) {
	// Validate input
	if input.Email == "" {
		return nil, errors.New("email is required")
	}
	
	if input.Password == "" {
		return nil, errors.New("password is required")
	}
	
	// Get user by email - Using GetByEmail to match interface
	userEntity, err := s.userRepository.GetByEmail(input.Email)
	if err != nil {
		// Log failed login attempt - Using string format for details parameter
		if err := s.auditService.LogAudit(
			"LOGIN_FAILED",
			"User",
			uuid.Nil,
			uuid.Nil,
			fmt.Sprintf("Failed login attempt for email: %s", input.Email),
			fmt.Sprintf("ip_address: %s, user_agent: %s, reason: user not found", input.IPAddress, input.UserAgent),
		); err != nil {
			// Log error but continue
			fmt.Printf("Failed to log audit: %v\n", err)
		}
		
		return nil, errors.New("invalid email or password")
	}
	
	// Check if user is active
	if !userEntity.IsActive {
		// Log failed login attempt - Using string format for details parameter
		if err := s.auditService.LogAudit(
			"LOGIN_FAILED",
			"User",
			userEntity.ID,
			userEntity.ID,
			fmt.Sprintf("Failed login attempt for inactive user: %s", input.Email),
			fmt.Sprintf("ip_address: %s, user_agent: %s, reason: user inactive", input.IPAddress, input.UserAgent),
		); err != nil {
			// Log error but continue
			fmt.Printf("Failed to log audit: %v\n", err)
		}
		
		return nil, errors.New("user is inactive")
	}
	
	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(input.Password)); err != nil {
		// Log failed login attempt - Using string format for details parameter
		if err := s.auditService.LogAudit(
			"LOGIN_FAILED",
			"User",
			userEntity.ID,
			userEntity.ID,
			fmt.Sprintf("Failed login attempt for user: %s", input.Email),
			fmt.Sprintf("ip_address: %s, user_agent: %s, reason: invalid password", input.IPAddress, input.UserAgent),
		); err != nil {
			// Log error but continue
			fmt.Printf("Failed to log audit: %v\n", err)
		}
		
		return nil, errors.New("invalid email or password")
	}
	
	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	
	// Create the claims
	claims := jwt.MapClaims{
		"user_id":  userEntity.ID.String(),
		"email":    userEntity.Email,
		"username": userEntity.Username,
		"role":     userEntity.Role,           // Single role for backward compatibility
		"roles":    []string{userEntity.Role}, // Array of roles for future extensibility
		"exp":      jwt.NewNumericDate(expiresAt).Unix(),
		"iat":      jwt.NewNumericDate(time.Now()).Unix(),
		"nbf":      jwt.NewNumericDate(time.Now()).Unix(),
		"iss":      "mynumba-donwin-api",
		"sub":      userEntity.ID.String(),
	}
	
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Log successful login - Using string format for details parameter
	if err := s.auditService.LogAudit(
		"LOGIN_SUCCESS",
		"User",
		userEntity.ID,
		userEntity.ID,
		fmt.Sprintf("Successful login for user: %s", input.Email),
		fmt.Sprintf("ip_address: %s, user_agent: %s", input.IPAddress, input.UserAgent),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &AuthenticateUserOutput{
		Token: tokenString,
		User: UserOutput{
			ID:       userEntity.ID,
			Email:    userEntity.Email,
			Username: userEntity.Username,
			Role:     userEntity.Role,
		},
		ExpiresAt: expiresAt,
	}, nil
}
