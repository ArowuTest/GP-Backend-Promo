package user

import (
	"context"
	"errors"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// UpdateUserInput represents the input for the UpdateUser use case
type UpdateUserInput struct {
	UserID    string
	Email     string
	FullName  string
	Role      string
	Active    *bool
	UpdatedBy string
}

// UpdateUserOutput represents the output from the UpdateUser use case
type UpdateUserOutput struct {
	User user.User
}

// UpdateUserUseCase defines the use case for updating a user
type UpdateUserUseCase struct {
	userRepo user.Repository
}

// NewUpdateUserUseCase creates a new UpdateUserUseCase
func NewUpdateUserUseCase(userRepo user.Repository) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo: userRepo,
	}
}

// Execute performs the update user use case
func (uc *UpdateUserUseCase) Execute(ctx context.Context, input UpdateUserInput) (UpdateUserOutput, error) {
	// Validate input
	if input.UserID == "" {
		return UpdateUserOutput{}, errors.New("user ID is required")
	}
	if input.UpdatedBy == "" {
		return UpdateUserOutput{}, errors.New("updater information is required")
	}

	// Get existing user
	existingUser, err := uc.userRepo.GetUserByID(ctx, input.UserID)
	if err != nil {
		return UpdateUserOutput{}, err
	}

	// Update user fields if provided
	if input.Email != "" {
		// Check if email already exists for another user
		if input.Email != existingUser.Email {
			exists, err := uc.userRepo.EmailExists(ctx, input.Email)
			if err != nil {
				return UpdateUserOutput{}, err
			}
			if exists {
				return UpdateUserOutput{}, errors.New("email already exists")
			}
			existingUser.Email = input.Email
		}
	}

	if input.FullName != "" {
		existingUser.FullName = input.FullName
	}

	if input.Role != "" {
		existingUser.Role = input.Role
	}

	if input.Active != nil {
		existingUser.Active = *input.Active
	}

	// Update metadata
	existingUser.UpdatedBy = input.UpdatedBy
	existingUser.UpdatedAt = time.Now()

	// Save user to repository
	updatedUser, err := uc.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		return UpdateUserOutput{}, err
	}

	return UpdateUserOutput{
		User: updatedUser,
	}, nil
}
