package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// UserServiceAdapter adapts the user service to a consistent interface
type UserServiceAdapter struct {
	// Internal services
	authenticateUserService *AuthenticateUserService
	createUserService       *CreateUserService
	updateUserService       *UpdateUserService
	getUserService          *GetUserService
	listUsersService        *ListUsersService
	resetPasswordService    *ResetPasswordService
}

// NewUserServiceAdapter creates a new UserServiceAdapter
func NewUserServiceAdapter(
	authenticateUserService *AuthenticateUserService,
	createUserService *CreateUserService,
	updateUserService *UpdateUserService,
	getUserService *GetUserService,
	listUsersService *ListUsersService,
	resetPasswordService *ResetPasswordService,
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

// AuthenticateUser authenticates a user
func (u *UserServiceAdapter) AuthenticateUser(
	ctx context.Context,
	email string,
	password string,
) (*entity.AuthResult, error) {
	// Create input for the service
	input := AuthenticateUserInput{
		Email:    email,
		Password: password,
	}

	// Authenticate user
	output, err := u.authenticateUserService.AuthenticateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.AuthResult{
		Token: output.Token,
		User: &entity.User{
			ID:        output.ID,
			Username:  output.Username,
			Email:     output.Email,
			Role:      output.Role,
			IsActive:  output.IsActive,
			CreatedAt: output.CreatedAt,
			UpdatedAt: output.UpdatedAt,
		},
		ExpiresAt: output.ExpiresAt,
	}

	return result, nil
}

// CreateUser creates a new user
func (u *UserServiceAdapter) CreateUser(
	ctx context.Context,
	email string,
	username string,
	password string,
	role string,
	createdBy uuid.UUID,
) (*entity.User, error) {
	// Create input for the service
	input := CreateUserInput{
		Email:     email,
		Username:  username,
		Password:  password,
		Role:      role,
		CreatedBy: createdBy,
	}

	// Create user
	output, err := u.createUserService.CreateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.User{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  output.IsActive,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}

	return result, nil
}

// UpdateUser updates a user
func (u *UserServiceAdapter) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	email string,
	username string,
	password string,
	role string,
	isActive bool,
	updatedBy uuid.UUID,
) (*entity.User, error) {
	// Create input for the service
	input := UpdateUserInput{
		ID:        id,
		Email:     email,
		Username:  username,
		Password:  password,
		Role:      role,
		IsActive:  isActive,
		UpdatedBy: updatedBy,
	}

	// Update user
	output, err := u.updateUserService.UpdateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.User{
		ID:        output.ID,
		Username:  output.Username,
		Email:     output.Email,
		Role:      output.Role,
		IsActive:  output.IsActive,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}

	return result, nil
}

// GetUserByID gets a user by ID
func (u *UserServiceAdapter) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.User, error) {
	// Create input for the service
	input := GetUserInput{
		ID: id,
	}

	// Get user
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
		IsActive:  output.IsActive,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}

	return result, nil
}

// ListUsers gets a list of users with pagination
func (u *UserServiceAdapter) ListUsers(
	ctx context.Context,
	page, pageSize int,
) (*entity.PaginatedUsers, error) {
	// Create input for the service
	input := ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	// Get users
	output, err := u.listUsersService.ListUsers(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create response
	result := &entity.PaginatedUsers{
		Users:      output.Users,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalCount: output.TotalCount,
		TotalPages: output.TotalPages,
	}

	return result, nil
}

// ResetPassword resets a user's password
func (u *UserServiceAdapter) ResetPassword(
	ctx context.Context,
	userID uuid.UUID,
	newPassword string,
	resetBy uuid.UUID,
) error {
	// Create input for the service
	input := ResetPasswordInput{
		UserID:      userID,
		NewPassword: newPassword,
		ResetBy:     resetBy,
	}

	// Reset password
	return u.resetPasswordService.ResetPassword(ctx, input)
}
