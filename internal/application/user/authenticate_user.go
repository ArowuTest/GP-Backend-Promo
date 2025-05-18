package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// AuthenticateUserService provides functionality for user authentication
type AuthenticateUserService struct {
	userRepository user.UserRepository
	auditService   audit.AuditService
}

// NewAuthenticateUserService creates a new AuthenticateUserService
func NewAuthenticateUserService(
	userRepository user.UserRepository,
	auditService audit.AuditService,
) *AuthenticateUserService {
	return &AuthenticateUserService{
		userRepository: userRepository,
		auditService:   auditService,
	}
}

// AuthenticateUserInput defines the input for the AuthenticateUser use case
type AuthenticateUserInput struct {
	Username string
	Password string
}

// AuthenticateUserOutput defines the output for the AuthenticateUser use case
type AuthenticateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	Token     string
	ExpiresAt time.Time
}

// AuthenticateUser authenticates a user and returns a JWT token
func (s *AuthenticateUserService) AuthenticateUser(ctx context.Context, input AuthenticateUserInput) (*AuthenticateUserOutput, error) {
	// Validate input
	if input.Username == "" {
		return nil, errors.New("username is required")
	}
	
	if input.Password == "" {
		return nil, errors.New("password is required")
	}
	
	// Get user by username
	user, err := s.userRepository.GetByUsername(input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		// Log failed login attempt
		if logErr := s.auditService.LogAudit(
			"LOGIN_FAILED",
			"User",
			user.ID,
			user.ID,
			fmt.Sprintf("Failed login attempt for user %s", user.Username),
			"Invalid password",
		); logErr != nil {
			// Log error but continue
			fmt.Printf("Failed to log audit: %v\n", logErr)
		}
		
		return nil, errors.New("invalid credentials")
	}
	
	// Generate JWT token
	token, expiresAt, err := generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Log successful login
	if err := s.auditService.LogAudit(
		"LOGIN_SUCCESS",
		"User",
		user.ID,
		user.ID,
		fmt.Sprintf("Successful login for user %s", user.Username),
		"",
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &AuthenticateUserOutput{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// generateJWTToken generates a JWT token for the user
func generateJWTToken(user *user.User) (string, time.Time, error) {
	// This is a simplified implementation
	// In a real-world scenario, this would use a JWT library
	
	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Generate a dummy token for demonstration
	token := fmt.Sprintf("dummy_token_%s_%s_%s", user.ID, user.Username, expiresAt.Format(time.RFC3339))
	
	return token, expiresAt, nil
}
