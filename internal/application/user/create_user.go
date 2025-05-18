package user

import (
	"context"
	"errors"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserInput represents the input for the CreateUser use case
type CreateUserInput struct {
	Username  string
	Password  string
	Email     string
	FullName  string
	Role      string
	CreatedBy string
}

// CreateUserOutput represents the output from the CreateUser use case
type CreateUserOutput struct {
	User user.User
}

// CreateUserUseCase defines the use case for creating a user
type CreateUserUseCase struct {
	userRepo user.Repository
}

// NewCreateUserUseCase creates a new CreateUserUseCase
func NewCreateUserUseCase(userRepo user.Repository) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
	}
}

// Execute performs the create user use case
func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (CreateUserOutput, error) {
	// Validate input
	if input.Username == "" {
		return CreateUserOutput{}, errors.New("username is required")
	}
	if input.Password == "" {
		return CreateUserOutput{}, errors.New("password is required")
	}
	if input.Email == "" {
		return CreateUserOutput{}, errors.New("email is required")
	}
	if input.Role == "" {
		return CreateUserOutput{}, errors.New("role is required")
	}
	if input.CreatedBy == "" {
		return CreateUserOutput{}, errors.New("creator information is required")
	}

	// Check if username already exists
	exists, err := uc.userRepo.UsernameExists(ctx, input.Username)
	if err != nil {
		return CreateUserOutput{}, err
	}
	if exists {
		return CreateUserOutput{}, errors.New("username already exists")
	}

	// Check if email already exists
	exists, err = uc.userRepo.EmailExists(ctx, input.Email)
	if err != nil {
		return CreateUserOutput{}, err
	}
	if exists {
		return CreateUserOutput{}, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return CreateUserOutput{}, err
	}

	// Create user entity
	newUser := user.User{
		Username:    input.Username,
		Password:    string(hashedPassword),
		Email:       input.Email,
		FullName:    input.FullName,
		Role:        input.Role,
		Active:      true,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Time{},
	}

	// Save user to repository
	createdUser, err := uc.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return CreateUserOutput{}, err
	}

	return CreateUserOutput{
		User: createdUser,
	}, nil
}
