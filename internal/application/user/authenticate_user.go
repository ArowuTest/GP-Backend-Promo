package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// AuthenticateUserService provides functionality for user authentication
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
	return &AuthenticateUserService{
		userRepository: userRepository,
		auditService:   auditService,
		jwtSecret:      "mynumba-donwin-jwt-secret-key-2025", // Using the same secret as in the middleware
	}
}

// AuthenticateUserInput defines the input for the AuthenticateUser use case
type AuthenticateUserInput struct {
	Username string
	Password string
	Email    string // Added Email field to support login by email
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

// Claims defines the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthenticateUser authenticates a user and returns a JWT token
func (s *AuthenticateUserService) AuthenticateUser(ctx context.Context, input AuthenticateUserInput) (*AuthenticateUserOutput, error) {
	// Validate input
	if input.Username == "" && input.Email == "" {
		return nil, errors.New("username or email is required")
	}
	
	if input.Password == "" {
		return nil, errors.New("password is required")
	}
	
	var user *user.User
	var err error
	
	// First try to get user by username
	if input.Username != "" {
		user, err = s.userRepository.GetByUsername(input.Username)
		if err != nil && input.Email != "" {
			// If username lookup fails and email is provided, try by email
			user, err = s.userRepository.GetByEmail(input.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to get user: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	} else if input.Email != "" {
		// If only email is provided, try by email
		user, err = s.userRepository.GetByEmail(input.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
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
	token, expiresAt, err := s.generateJWTToken(user)
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
func (s *AuthenticateUserService) generateJWTToken(user *user.User) (string, time.Time, error) {
	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Create claims with user information
	claims := &Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mynumba-donwin-api",
			Subject:   user.ID.String(),
		},
	}
	
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	
	return tokenString, expiresAt, nil
}
