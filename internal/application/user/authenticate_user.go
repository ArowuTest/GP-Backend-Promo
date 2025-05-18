package application

import (
	"time"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// AuthenticateUserUseCase represents the use case for authenticating a user
type AuthenticateUserUseCase struct {
	userRepository user.UserRepository
}

// NewAuthenticateUserUseCase creates a new AuthenticateUserUseCase
func NewAuthenticateUserUseCase(
	userRepository user.UserRepository,
) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{
		userRepository: userRepository,
	}
}

// AuthenticateUserInput represents the input for the authenticate user use case
type AuthenticateUserInput struct {
	Email    string
	Password string
	IPAddress string
	UserAgent string
}

// AuthenticateUserOutput represents the output of the authenticate user use case
type AuthenticateUserOutput struct {
	User  *user.User
	Token string
}

// Execute authenticates a user with the provided credentials
func (uc *AuthenticateUserUseCase) Execute(input AuthenticateUserInput) (*AuthenticateUserOutput, error) {
	// Verify credentials
	authenticatedUser, err := uc.userRepository.VerifyCredentials(input.Email, input.Password)
	if err != nil {
		return nil, user.NewUserError(user.ErrInvalidCredentials, "Invalid email or password", err)
	}

	// Update last login time
	now := time.Now()
	authenticatedUser.LastLogin = &now
	authenticatedUser.UpdatedAt = now

	if err := uc.userRepository.Update(authenticatedUser); err != nil {
		// Non-critical error, we can continue even if this fails
		// Just log the error in a real implementation
	}

	// Generate JWT token (simplified for example)
	token := "jwt_token_would_be_generated_here"

	return &AuthenticateUserOutput{
		User:  authenticatedUser,
		Token: token,
	}, nil
}
