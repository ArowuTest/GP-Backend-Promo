package admin

import (
	"errors"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAdminUserRequest defines the expected request body for creating an admin user
type CreateAdminUserRequest struct {
	Username  string               `json:"username" binding:"required"`
	Email     string               `json:"email" binding:"required,email"`
	Password  string               `json:"password" binding:"required,min=8"`
	FirstName string               `json:"first_name,omitempty"`
	LastName  string               `json:"last_name,omitempty"`
	Role      models.AdminUserRole `json:"role" binding:"required"`
	Status    models.UserStatus    `json:"status,omitempty"` // e.g., Active, Inactive
}

// AdminUserResponse defines the structure for admin user responses, omitting sensitive data
type AdminUserResponse struct {
	ID        uuid.UUID            `json:"id"`
	Username  string               `json:"username"`
	Email     string               `json:"email"`
	FirstName string               `json:"first_name,omitempty"`
	LastName  string               `json:"last_name,omitempty"`
	Role      models.AdminUserRole `json:"role"`
	Status    models.UserStatus    `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	LastLoginAt *time.Time         `json:"last_login_at,omitempty"`
}

// toAdminUserResponse converts an AdminUser model to an AdminUserResponse
func toAdminUserResponse(user *models.AdminUser) AdminUserResponse {
	return AdminUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}

// CreateAdminUser godoc
// @Summary Create a new admin user
// @Description Create a new admin user with username, password, email, first/last name, and role.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param user body CreateAdminUserRequest true "AdminUser object to be created."
// @Success 201 {object} AdminUserResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [post]
func CreateAdminUser(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	switch req.Role {
	case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	if req.Status == "" {
		req.Status = models.StatusActive // Default to Active if not provided
	} else {
	    switch req.Status {
	    case models.StatusActive, models.StatusInactive, models.StatusLocked:
	    default:
	        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user status specified"})
	        return
	    }
	}

	salt, err := auth.GenerateSalt()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate salt: " + err.Error()})
        return
    }
	hashedPassword, err := auth.HashPassword(req.Password, salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password: " + err.Error()})
		return
	}

	newUser := models.AdminUser{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Salt:         salt,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		Status:       req.Status,
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toAdminUserResponse(&newUser))
}

// GetAdminUser godoc
// @Summary Get an admin user by ID
// @Description Get details of a specific admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} AdminUserResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Router /admin/users/{id} [get]
func GetAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user models.AdminUser
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
		    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, toAdminUserResponse(&user))
}

// UpdateAdminUserRequest defines the expected request body for updating an admin user
type UpdateAdminUserRequest struct {
	Username  *string               `json:"username,omitempty"`
	Email     *string               `json:"email,omitempty"` // Add email validation if allowing update
	Password  *string               `json:"password,omitempty"` // Min 8 if provided
	FirstName *string               `json:"first_name,omitempty"`
	LastName  *string               `json:"last_name,omitempty"`
	Role      *models.AdminUserRole `json:"role,omitempty"`
	Status    *models.UserStatus    `json:"status,omitempty"`
}

// UpdateAdminUser godoc
// @Summary Update an existing admin user
// @Description Update details of an existing admin user.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body UpdateAdminUserRequest true "AdminUser object with updated fields"
// @Success 200 {object} AdminUserResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [put]
func UpdateAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var existingUser models.AdminUser
	if err := config.DB.First(&existingUser, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user for update: " + err.Error()})
        }
		return
	}

	var req UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Role != nil {
		switch *req.Role {
		case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
			updates["role"] = *req.Role
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
			return
		}
	}
	if req.Status != nil {
	    switch *req.Status {
	    case models.StatusActive, models.StatusInactive, models.StatusLocked:
	        updates["status"] = *req.Status
	    default:
	        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user status specified"})
	        return
	    }
	}

	if req.Password != nil && *req.Password != "" {
	    if len(*req.Password) < 8 {
	        c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 8 characters long"})
	        return
	    }
	    // Salt should be regenerated or fetched if we are only storing hash. For simplicity, using existing salt.
	    // If salt is unique per user and stored, it should be used. If salt is global, it can be fetched.
	    // The current model has a Salt field, implying it is per-user.
	    // If we want to update password, we should ideally re-salt or ensure the existing salt is used correctly.
        // For this implementation, we will use the existing salt.
		hashedPassword, err := auth.HashPassword(*req.Password, existingUser.Salt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password: " + err.Error()})
			return
		}
		updates["password_hash"] = hashedPassword
	}

    if len(updates) > 0 {
        if err := config.DB.Model(&existingUser).Updates(updates).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
            return
        }
    }

	c.JSON(http.StatusOK, toAdminUserResponse(&existingUser))
}

// DeleteAdminUser godoc
// @Summary Delete an admin user by ID (Soft Delete)
// @Description Soft delete an admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} gin.H{"message": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [delete]
func DeleteAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Prevent deletion of the primary superadmin or self-deletion if needed (add logic here)

	result := config.DB.Delete(&models.AdminUser{}, "id = ?", userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
	    c.JSON(http.StatusNotFound, gin.H{"error": "User not found or already deleted"})
	    return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListAdminUsers godoc
// @Summary List all admin users
// @Description Get a list of all admin users.
// @Tags AdminUsers
// @Produce json
// @Success 200 {array} AdminUserResponse
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [get]
func ListAdminUsers(c *gin.Context) {
	var users []models.AdminUser
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
		return
	}

	responseUsers := make([]AdminUserResponse, len(users))
	for i, user := range users {
		responseUsers[i] = toAdminUserResponse(&user)
	}
	c.JSON(http.StatusOK, responseUsers)
}

// LoginRequest defines the structure for login requests
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse defines the structure for login responses
type LoginResponse struct {
	Token    string               `json:"token"`
	UserID   uuid.UUID            `json:"user_id"`
	Username string               `json:"username"`
	Role     models.AdminUserRole `json:"role"`
}

// Login godoc
// @Summary Admin user login
// @Description Authenticate an admin user and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post]
func Login(c *gin.Context) {
	var creds LoginRequest

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var user models.AdminUser
	if err := config.DB.Where("username = ? AND status = ?", creds.Username, models.StatusActive).First(&user).Error; err != nil {
	    if errors.Is(err, gorm.ErrRecordNotFound) {
	        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username, password, or user is not active"})
	    } else {
	        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login: " + err.Error()})
	    }
		return
	}

	if !auth.CheckPasswordHash(creds.Password, user.Salt, user.PasswordHash) {
	    // Log failed attempt
	    user.FailedLoginAttempts++
	    if user.FailedLoginAttempts >= 5 { // Example: Lock after 5 attempts
	        user.Status = models.StatusLocked
	    }
	    config.DB.Save(&user)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

    // Reset failed attempts and update last login on successful login
    user.FailedLoginAttempts = 0
    now := time.Now()
    user.LastLoginAt = &now
    config.DB.Save(&user)

	token, err := auth.GenerateJWT(user.ID.String(), user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token, UserID: user.ID, Username: user.Username, Role: user.Role})
}

