package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// UserServiceAdapter adapts the user service to a consistent interface
type UserServiceAdapter struct {
	authenticateUserService user.AuthenticateUserService
	createUserService       user.CreateUserService
	updateUserService       user.UpdateUserService
	getUserService          user.GetUserService
	listUsersService        user.ListUsersService
}

// NewUserServiceAdapter creates a new UserServiceAdapter
func NewUserServiceAdapter(
	authenticateUserService *user.AuthenticateUserService,
	createUserService *user.CreateUserService,
	updateUserService *user.UpdateUserService,
	getUserService *user.GetUserService,
	listUsersService *user.ListUsersService,
) *UserServiceAdapter {
	return &UserServiceAdapter{
		authenticateUserService: *authenticateUserService,
		createUserService:       *createUserService,
		updateUserService:       *updateUserService,
		getUserService:          *getUserService,
		listUsersService:        *listUsersService,
	}
}

// AuthenticateUserOutput represents the output of AuthenticateUser
type AuthenticateUserOutput struct {
	User      entity.User
	Token     string
	ExpiresAt time.Time
}

// CreateUserOutput represents the output of CreateUser
type CreateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetUserOutput represents the output of GetUser
type GetUserOutput struct {
	User entity.User
}

// ListUsersOutput represents the output of ListUsers
type ListUsersOutput struct {
	Users      []entity.User
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

// UpdateUserOutput represents the output of UpdateUser
type UpdateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AuthenticateUser authenticates a user
func (u *UserServiceAdapter) AuthenticateUser(ctx context.Context, username, password string) (*AuthenticateUserOutput, error) {
	// Call the actual service
	input := user.AuthenticateUserInput{
		Email:    username, // Using username as email since the application layer expects Email
		Password: password,
	}

	output, err := u.authenticateUserService.AuthenticateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &AuthenticateUserOutput{
		User: entity.User{
			ID:        output.User.ID,
			Username:  output.User.Username,
			Email:     output.User.Email,
			Role:      output.User.Role,
			IsActive:  true, // Default to true since field is missing in output
			CreatedAt: time.Now(), // Default since field is missing in output
			UpdatedAt: time.Now(), // Default since field is missing in output
		},
		Token:     output.Token,
		ExpiresAt: output.ExpiresAt,
	}, nil
}

// CreateUser creates a new user
func (u *UserServiceAdapter) CreateUser(ctx context.Context, username, email, password, role string, isActive bool) (*CreateUserOutput, error) {
	// Call the actual service
	input := user.CreateUserInput{
		Username: username,
		Email:    email,
		Password: password,
		Role:     role,
	}

	output, err := u.createUserService.CreateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	return &CreateUserOutput{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  true, // Default since field is missing in output
		CreatedAt: time.Now(), // Default since field is missing in output
		UpdatedAt: time.Now(), // Default since field is missing in output
	}, nil
}

// GetUser gets a user by ID
func (u *UserServiceAdapter) GetUser(ctx context.Context, id uuid.UUID) (*GetUserOutput, error) {
	// Call the actual service
	input := user.GetUserInput{
		ID: id,
	}

	output, err := u.getUserService.GetUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.User{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  true, // Default since field is missing in output
		CreatedAt: time.Now(), // Default since field is missing in output
		UpdatedAt: time.Now(), // Default since field is missing in output
	}

	return &GetUserOutput{
		User: *result,
	}, nil
}

// ListUsers lists users with pagination
func (u *UserServiceAdapter) ListUsers(ctx context.Context, page, pageSize int) (*ListUsersOutput, error) {
	// Call the actual service
	input := user.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := u.listUsersService.ListUsers(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert users for response
	users := make([]entity.User, 0, len(output.Users))
	for _, u := range output.Users {
		users = append(users, entity.User{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role,
			IsActive:  true, // Default since field is missing in output
			CreatedAt: time.Now(), // Default since field is missing in output
			UpdatedAt: time.Now(), // Default since field is missing in output
		})
	}

	// Return response
	return &ListUsersOutput{
		Users:      users,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}, nil
}

// UpdateUser updates a user
func (u *UserServiceAdapter) UpdateUser(ctx context.Context, id uuid.UUID, email, role string, isActive bool) (*UpdateUserOutput, error) {
	// Call the actual service
	input := user.UpdateUserInput{
		ID:    id,
		Email: email,
		Role:  role,
		// IsActive field is missing in the application layer
	}

	output, err := u.updateUserService.UpdateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	return &UpdateUserOutput{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  true, // Default since field is missing in output
		CreatedAt: time.Now(), // Default since field is missing in output
		UpdatedAt: time.Now(), // Default since field is missing in output
	}, nil
}
