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

// UpdateUserService provides functionality for updating users
type UpdateUserService struct {
	userRepository user.UserRepository
	auditService   audit.AuditService
}

// NewUpdateUserService creates a new UpdateUserService
func NewUpdateUserService(
	userRepository user.UserRepository,
	auditService audit.AuditService,
) *UpdateUserService {
	return &UpdateUserService{
		userRepository: userRepository,
		auditService:   auditService,
	}
}

// UpdateUserInput defines the input for the UpdateUser use case
type UpdateUserInput struct {
	ID        uuid.UUID
	Email     string
	Username  string
	Password  string // Optional, if empty, password won't be updated
	Role      string
	IsActive  bool
	UpdatedBy uuid.UUID
}

// UpdateUserOutput defines the output for the UpdateUser use case
type UpdateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpdateUser updates an existing user
func (s *UpdateUserService) UpdateUser(ctx context.Context, input UpdateUserInput) (*UpdateUserOutput, error) {
	// Validate input
	if input.ID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	
	if input.Email == "" {
		return nil, errors.New("email is required")
	}
	
	if input.Role == "" {
		return nil, errors.New("role is required")
	}
	
	// Get existing user
	user, err := s.userRepository.GetByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Update user fields
	user.Email = input.Email
	if input.Username != "" {
		user.Username = input.Username
	}
	user.Role = input.Role
	user.IsActive = input.IsActive
	user.UpdatedAt = time.Now()
	
	// Update password if provided
	if input.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(passwordHash)
	}
	
	// Save user
	if err := s.userRepository.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	// Log audit
	if err := s.auditService.LogAudit(
		"UPDATE_USER",
		"User",
		user.ID,
		input.UpdatedBy,
		fmt.Sprintf("User updated: %s", user.Username),
		fmt.Sprintf("Role: %s", user.Role),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &UpdateUserOutput{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
