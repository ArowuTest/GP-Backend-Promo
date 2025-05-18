package user

import (
	"context"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// ListUsersInput represents the input for the ListUsers use case
type ListUsersInput struct {
	Page     int
	PageSize int
}

// ListUsersOutput represents the output from the ListUsers use case
type ListUsersOutput struct {
	Users    []user.User
	Total    int64
	Page     int
	PageSize int
}

// ListUsersUseCase defines the use case for listing users
type ListUsersUseCase struct {
	userRepo user.Repository
}

// NewListUsersUseCase creates a new ListUsersUseCase
func NewListUsersUseCase(userRepo user.Repository) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute performs the list users use case
func (uc *ListUsersUseCase) Execute(ctx context.Context, input ListUsersInput) (ListUsersOutput, error) {
	// Set default page size if not provided
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	// Set default page if not provided
	if input.Page <= 0 {
		input.Page = 1
	}

	// Prepare filter criteria
	filter := user.UserFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	// Get users from repository
	users, err := uc.userRepo.ListUsers(ctx, filter)
	if err != nil {
		return ListUsersOutput{}, err
	}

	// Get total count for pagination
	total, err := uc.userRepo.CountUsers(ctx, filter)
	if err != nil {
		return ListUsersOutput{}, err
	}

	return ListUsersOutput{
		Users:    users,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}
