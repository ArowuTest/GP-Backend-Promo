package user

import (
	"context"
	"fmt"
	
	userDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// ListUsersService provides functionality for listing users
type ListUsersService struct {
	userRepository userDomain.UserRepository
}

// NewListUsersService creates a new ListUsersService
func NewListUsersService(userRepository userDomain.UserRepository) *ListUsersService {
	return &ListUsersService{
		userRepository: userRepository,
	}
}

// ListUsersInput defines the input for the ListUsers use case
type ListUsersInput struct {
	Page     int
	PageSize int
}

// ListUsersOutput defines the output for the ListUsers use case
type ListUsersOutput struct {
	Users      []userDomain.User
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ListUsers retrieves a paginated list of users
func (s *ListUsersService) ListUsers(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	
	users, totalCount, err := s.userRepository.List(input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}
	
	return &ListUsersOutput{
		Users:      users,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
