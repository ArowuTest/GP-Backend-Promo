package gorm

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/user"
)

// GormUserRepository implements the user.UserRepository interface using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{
		db: db,
	}
}

// UserModel is the GORM model for users
type UserModel struct {
	ID           string     `gorm:"primaryKey;type:uuid"`
	Email        string     `gorm:"uniqueIndex"`
	Username     string
	FullName     string
	Role         string
	PasswordHash string
	LastLogin    *time.Time
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName returns the table name for the UserModel
func (UserModel) TableName() string {
	return "users"
}

// toModel converts a domain user entity to a GORM model
func toUserModel(u *user.User) *UserModel {
	return &UserModel{
		ID:           u.ID.String(),
		Email:        u.Email,
		Username:     u.Username,
		FullName:     u.FullName,
		Role:         u.Role,
		PasswordHash: u.PasswordHash,
		LastLogin:    u.LastLogin,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain user entity
func (m *UserModel) toDomain() (*user.User, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	return &user.User{
		ID:           id,
		Email:        m.Email,
		Username:     m.Username,
		FullName:     m.FullName,
		Role:         m.Role,
		PasswordHash: m.PasswordHash,
		LastLogin:    m.LastLogin,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

// Create implements the user.UserRepository interface
func (r *GormUserRepository) Create(u *user.User) error {
	model := toUserModel(u)
	result := r.db.Create(model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return user.NewUserError(user.ErrEmailAlreadyExists, "Email already exists", result.Error)
		}
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	
	return nil
}

// GetByID implements the user.UserRepository interface
func (r *GormUserRepository) GetByID(id uuid.UUID) (*user.User, error) {
	var model UserModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.NewUserError(user.ErrUserNotFound, "User not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	
	userEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
	}
	
	return userEntity, nil
}

// GetByEmail implements the user.UserRepository interface
func (r *GormUserRepository) GetByEmail(email string) (*user.User, error) {
	var model UserModel
	result := r.db.Where("email = ?", email).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.NewUserError(user.ErrUserNotFound, "User not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	
	userEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
	}
	
	return userEntity, nil
}

// GetByUsername implements the user.UserRepository interface
func (r *GormUserRepository) GetByUsername(username string) (*user.User, error) {
	var model UserModel
	result := r.db.Where("username = ?", username).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.NewUserError(user.ErrUserNotFound, "User not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	
	userEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
	}
	
	return userEntity, nil
}

// List implements the user.UserRepository interface
func (r *GormUserRepository) List(page, pageSize int) ([]user.User, int, error) {
	var models []UserModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Get total count
	result := r.db.Model(&UserModel{}).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", result.Error)
	}
	
	// Get paginated users
	result = r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", result.Error)
	}
	
	users := make([]user.User, 0, len(models))
	for _, model := range models {
		userEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert user model to domain: %w", err)
		}
		users = append(users, *userEntity)
	}
	
	return users, int(total), nil
}

// Update implements the user.UserRepository interface
func (r *GormUserRepository) Update(u *user.User) error {
	model := toUserModel(u)
	result := r.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	
	return nil
}

// Delete implements the user.UserRepository interface
func (r *GormUserRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&UserModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	
	return nil
}

// VerifyCredentials implements the user.UserRepository interface
func (r *GormUserRepository) VerifyCredentials(email, password string) (*user.User, error) {
	var model UserModel
	result := r.db.Where("email = ?", email).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.NewUserError(user.ErrInvalidCredentials, "Invalid credentials", result.Error)
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	
	// Check if user is active
	if !model.IsActive {
		return nil, user.NewUserError(user.ErrInvalidCredentials, "User is inactive", nil)
	}
	
	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(model.PasswordHash), []byte(password))
	if err != nil {
		return nil, user.NewUserError(user.ErrInvalidCredentials, "Invalid credentials", err)
	}
	
	userEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
	}
	
	return userEntity, nil
}
