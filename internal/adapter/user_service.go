package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
)

// UserServiceAdapter adapts the user service to a consistent interface
type UserServiceAdapter struct {
	authenticateUserService user.AuthenticateUserService
	createUserService       user.CreateUserService
	updateUserService       user.UpdateUserService
	getUserService          user.GetUserService
	listUsersService        user.ListUsersService
	resetPasswordService    user.ResetPasswordService
}

// NewUserServiceAdapter creates a new UserServiceAdapter
func NewUserServiceAdapter(
	authenticateUserService user.AuthenticateUserService,
	createUserService user.CreateUserService,
	updateUserService user.UpdateUserService,
	getUserService user.GetUserService,
	listUsersService user.ListUsersService,
	resetPasswordService user.ResetPasswordService,
) *UserServiceAdapter {
	return &UserServiceAdapter{
		authenticateUserService: authenticateUserService,
		createUserService:       createUserService,
		updateUserService:       updateUserService,
		getUserService:          getUserService,
		listUsersService:        listUsersService,
		resetPasswordService:    resetPasswordService,
	}
}

// User represents a user
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AuthenticateUserOutput represents the output of AuthenticateUser
type AuthenticateUserOutput struct {
	Token     string
	User      User
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

// UpdateUserOutput represents the output of UpdateUser
type UpdateUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	UpdatedAt time.Time
}

// GetUserOutput represents the output of GetUser
type GetUserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListUsersOutput represents the output of ListUsers
type ListUsersOutput struct {
	Users       []User
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

// AuthenticateUser authenticates a user
func (u *UserServiceAdapter) AuthenticateUser(ctx context.Context, email, password string) (*AuthenticateUserOutput, error) {
	// Call the actual service
	input := user.AuthenticateUserInput{
		Email:    email,
		Password: password,
	}

	output, err := u.authenticateUserService.AuthenticateUser(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &AuthenticateUserOutput{
		Token: output.Token,
		User: User{
			ID:        output.User.ID,
			Username:  output.User.Username,
			Email:     output.User.Email,
			Role:      output.User.Role,
			IsActive:  output.User.IsActive,
			CreatedAt: output.User.CreatedAt,
			UpdatedAt: output.User.UpdatedAt,
		},
		ExpiresAt: output.ExpiresAt,
	}, nil
}

// CreateUser creates a user
func (u *UserServiceAdapter) CreateUser(ctx context.Context, email, username, password, role string, createdBy uuid.UUID) (*CreateUserOutput, error) {
	// Call the actual service
	input := user.CreateUserInput{
		Email:     email,
		Username:  username,
		Password:  password,
		Role:      role,
		CreatedBy: createdBy,
	}

	output, err := u.createUserService.CreateUser(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &CreateUserOutput{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  output.IsActive,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}, nil
}

// UpdateUser updates a user
func (u *UserServiceAdapter) UpdateUser(ctx context.Context, id uuid.UUID, email, username, password, role string, isActive bool, updatedBy uuid.UUID) (*UpdateUserOutput, error) {
	// Call the actual service
	input := user.UpdateUserInput{
		ID:        id,
		Email:     email,
		Username:  username,
		Password:  password,
		Role:      role,
		IsActive:  isActive,
		UpdatedBy: updatedBy,
	}

	output, err := u.updateUserService.UpdateUser(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &UpdateUserOutput{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  output.IsActive,
		UpdatedAt: output.UpdatedAt,
	}, nil
}

// GetUser gets a user by ID
func (u *UserServiceAdapter) GetUser(ctx context.Context, id uuid.UUID) (*GetUserOutput, error) {
	// Call the actual service
	output, err := u.getUserService.GetUser(id)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	return &GetUserOutput{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  output.IsActive,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}, nil
}

// ListUsers lists users with pagination
func (u *UserServiceAdapter) ListUsers(ctx context.Context, page, pageSize int) (*ListUsersOutput, error) {
	// Call the actual service
	input := user.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := u.listUsersService.ListUsers(input)
	if err != nil {
		return nil, err
	}

	// Convert to adapter output
	users := make([]User, 0, len(output.Users))
	for _, user := range output.Users {
		users = append(users, User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return &ListUsersOutput{
		Users:       users,
		Page:        output.Page,
		PageSize:    output.PageSize,
		TotalCount:  output.TotalCount,
		TotalPages:  output.TotalPages,
	}, nil
}

// ResetPassword resets a user's password
func (u *UserServiceAdapter) ResetPassword(ctx context.Context, email, oldPassword, newPassword string, resetBy uuid.UUID) error {
	// Call the actual service
	input := user.ResetPasswordInput{
		Email:       email,
		OldPassword: oldPassword,
		NewPassword: newPassword,
		ResetBy:     resetBy,
	}

	return u.resetPasswordService.ResetPassword(input)
}
