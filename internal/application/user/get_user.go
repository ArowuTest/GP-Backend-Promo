package user

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	userDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// GetUserService provides functionality for retrieving users
type GetUserService struct {
	userRepository userDomain.UserRepository
}

// NewGetUserService creates a new GetUserService
func NewGetUserService(userRepository userDomain.UserRepository) *GetUserService {
	return &GetUserService{
		userRepository: userRepository,
	}
}

// GetUserInput defines the input for the GetUser use case
type GetUserInput struct {
	ID uuid.UUID
}

// GetUserOutput defines the output for the GetUser use case
type GetUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetUser retrieves a user by ID
func (s *GetUserService) GetUser(ctx context.Context, input GetUserInput) (*GetUserOutput, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	
	user, err := s.userRepository.GetByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &GetUserOutput{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
