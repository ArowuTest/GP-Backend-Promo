package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
	
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// ResetPasswordService provides functionality for resetting user passwords
type ResetPasswordService struct {
	userRepository user.UserRepository
	auditService   audit.AuditService
}

// NewResetPasswordService creates a new ResetPasswordService
func NewResetPasswordService(
	userRepository user.UserRepository,
	auditService audit.AuditService,
) *ResetPasswordService {
	return &ResetPasswordService{
		userRepository: userRepository,
		auditService:   auditService,
	}
}

// ResetPasswordInput defines the input for the ResetPassword use case
type ResetPasswordInput struct {
	UserID       uuid.UUID
	NewPassword  string
	AdminUserID  uuid.UUID // ID of the admin performing the reset
}

// ResetPasswordOutput defines the output for the ResetPassword use case
type ResetPasswordOutput struct {
	UserID    uuid.UUID
	Username  string
	Email     string
	UpdatedAt time.Time
}

// ResetPassword resets a user's password
func (s *ResetPasswordService) ResetPassword(ctx context.Context, input ResetPasswordInput) (*ResetPasswordOutput, error) {
	// Validate input
	if input.UserID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	
	if input.NewPassword == "" {
		return nil, errors.New("new password is required")
	}
	
	// Validate password strength
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return nil, err
	}
	
	// Get user by ID
	user, err := s.userRepository.GetByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Generate bcrypt hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Update user password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	
	// Save user
	err = s.userRepository.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	// Log password reset
	err = s.auditService.LogAudit(
		"PASSWORD_RESET",
		"User",
		user.ID,
		input.AdminUserID,
		fmt.Sprintf("Password reset for user %s by admin", user.Username),
		"",
	)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &ResetPasswordOutput{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// validatePasswordStrength validates that a password meets strength requirements
func validatePasswordStrength(password string) error {
	// Check length
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	// Check for uppercase
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	
	// Check for lowercase
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	
	// Check for number
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one number")
	}
	
	// Check for special character
	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}

// IsAdminUser checks if the user has admin privileges
func IsAdminUser(ctx context.Context, userRepository user.UserRepository) (bool, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return false, errors.New("user not authenticated")
	}
	
	user, err := userRepository.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user.Role == "SUPER_ADMIN" || user.Role == "ADMIN", nil
}

// GetCurrentUserID gets the current user ID from context
func GetCurrentUserID(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}
