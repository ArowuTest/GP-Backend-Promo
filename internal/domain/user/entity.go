package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Role      string // "SuperAdmin", "Admin", "SeniorUser", "WinnersReportUser", "AllReportUser"
	Password  string // Hashed password
	LastLogin *time.Time
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	List(page, pageSize int) ([]User, int, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	VerifyCredentials(email, password string) (*User, error)
}

// UserError represents domain-specific errors for the user domain
type UserError struct {
	Code    string
	Message string
	Err     error
}

// Error codes for the user domain
const (
	ErrUserNotFound      = "USER_NOT_FOUND"
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	ErrInvalidEmail      = "INVALID_EMAIL"
	ErrInvalidPassword   = "INVALID_PASSWORD"
	ErrInvalidRole       = "INVALID_ROLE"
)

// Error implements the error interface
func (e *UserError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a new UserError
func NewUserError(code, message string, err error) *UserError {
	return &UserError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidateUser validates that a user is valid
func ValidateUser(user *User) error {
	if user.Email == "" {
		return errors.New("email cannot be empty")
	}
	
	if user.Name == "" {
		return errors.New("name cannot be empty")
	}
	
	if user.Role == "" {
		return errors.New("role cannot be empty")
	}
	
	// Validate role is one of the allowed values
	validRoles := map[string]bool{
		"SuperAdmin":       true,
		"Admin":            true,
		"SeniorUser":       true,
		"WinnersReportUser": true,
		"AllReportUser":    true,
	}
	
	if !validRoles[user.Role] {
		return errors.New("invalid role")
	}
	
	return nil
}

// ValidatePassword validates that a password meets the required criteria
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	// Additional password validation logic can be added here
	
	return nil
}
