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

// CreateUserService provides functionality for creating users
type CreateUserService struct {
	userRepository user.UserRepository
	auditService   audit.AuditService
}

// NewCreateUserService creates a new CreateUserService
func NewCreateUserService(
	userRepository user.UserRepository,
	auditService audit.AuditService,
) *CreateUserService {
	return &CreateUserService{
		userRepository: userRepository,
		auditService:   auditService,
	}
}

// CreateUserInput defines the input for the CreateUser use case
type CreateUserInput struct {
	Username  string
	Email     string
	Password  string
	Role      string
	CreatedBy uuid.UUID
}

// CreateUserOutput defines the output for the CreateUser use case
type CreateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	CreatedAt time.Time
}

// CreateUser creates a new user
func (s *CreateUserService) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// Validate input
	if input.Username == "" {
		return nil, errors.New("username is required")
	}
	
	if input.Email == "" {
		return nil, errors.New("email is required")
	}
	
	if input.Password == "" {
		return nil, errors.New("password is required")
	}
	
	if input.Role == "" {
		return nil, errors.New("role is required")
	}
	
	// Check if username already exists
	existingUser, err := s.userRepository.GetByUsername(input.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}
	
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create user
	now := time.Now()
	user := &user.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Role:         input.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	if err := s.userRepository.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Log audit
	if err := s.auditService.LogAudit(
		"CREATE_USER",
		"User",
		user.ID,
		input.CreatedBy,
		fmt.Sprintf("User created: %s", input.Username),
		fmt.Sprintf("Role: %s", input.Role),
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &CreateUserOutput{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}
