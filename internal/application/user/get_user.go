package user

import (
	"context"
	"errors"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// GetUserInput represents the input for the GetUser use case
type GetUserInput struct {
	UserID string
}

// GetUserOutput represents the output from the GetUser use case
type GetUserOutput struct {
	User user.User
}

// GetUserUseCase defines the use case for retrieving a user
type GetUserUseCase struct {
	userRepo user.Repository
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userRepo user.Repository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

// Execute performs the get user use case
func (uc *GetUserUseCase) Execute(ctx context.Context, input GetUserInput) (GetUserOutput, error) {
	// Validate input
	if input.UserID == "" {
		return GetUserOutput{}, errors.New("user ID is required")
	}

	// Get user from repository
	user, err := uc.userRepo.GetUserByID(ctx, input.UserID)
	if err != nil {
		return GetUserOutput{}, err
	}

	return GetUserOutput{
		User: user,
	}, nil
}
