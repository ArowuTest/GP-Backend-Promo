package admin

import (
	"net/http"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateAdminUserRequest struct for creating a new admin user
type CreateAdminUserRequest struct {
	Username  string                `json:"username" binding:"required"`
	Email     string                `json:"email" binding:"required,email"`
	Password  string                `json:"password" binding:"required,min=8"`
	FirstName string                `json:"first_name,omitempty"`
	LastName  string                `json:"last_name,omitempty"`
	Role      models.AdminUserRole  `json:"role" binding:"required"`
	Status    models.UserStatus     `json:"status,omitempty"` // Default to Active in model
}

// CreateAdminUser handles the creation of a new admin user (SuperAdmin only)
func CreateAdminUser(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Check if username or email already exists
	var existingUserByUsername models.AdminUser
	if !errors.Is(config.DB.Where("username = ?", req.Username).First(&existingUserByUsername).Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}
	var existingUserByEmail models.AdminUser
	if !errors.Is(config.DB.Where("email = ?", req.Email).First(&existingUserByEmail).Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	hashedPassword, saltValue, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := models.AdminUser{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Salt:         saltValue,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		Status:       models.StatusActive, // Default to active
	}
	if req.Status != "" {
		newUser.Status = req.Status
	}

	result := config.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// ListAdminUsers handles listing all admin users (SuperAdmin only)
func ListAdminUsers(c *gin.Context) {
	var users []models.AdminUser
	result := config.DB.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin users: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetAdminUser handles retrieving a single admin user by ID (SuperAdmin only)
func GetAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user models.AdminUser
	result := config.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin user: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateAdminUserRequest struct for updating an admin user
type UpdateAdminUserRequest struct {
	FirstName *string               `json:"first_name,omitempty"`
	LastName  *string               `json:"last_name,omitempty"`
	Role      *models.AdminUserRole `json:"role,omitempty"`
	Status    *models.UserStatus    `json:"status,omitempty"`
	Password  *string               `json:"password,omitempty"` // Allow password update
}

// UpdateAdminUser handles updating an admin user (SuperAdmin only)
func UpdateAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var user models.AdminUser
	if config.DB.First(&user, "id = ?", userID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	// Prepare map for updates to handle partial updates and avoid zero values overwriting existing data unintentionally
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		// Validate status value
		if *req.Status != models.StatusActive && *req.Status != models.StatusInactive && *req.Status != models.StatusLocked {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
			return
		}
		updates["status"] = *req.Status
	}

	if req.Password != nil && *req.Password != "" {
		hashedPassword, saltValue, err := auth.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		updates["password_hash"] = hashedPassword
		updates["salt"] = saltValue
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided"})
		return
	}

	result := config.DB.Model(&user).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin user: " + result.Error.Error()})
		return
	}

	// Refetch to get updated data
	config.DB.First(&user, "id = ?", userID)
	c.JSON(http.StatusOK, user)
}

// DeleteAdminUser handles deleting an admin user (SuperAdmin only) - Soft delete is used
func DeleteAdminUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Prevent SuperAdmin from deleting their own account (especially if only one)
	// This logic should be more robust, e.g., check if it's the last SuperAdmin
	loggedInUserIDStr, exists := c.Get("userID")
	if exists && loggedInUserIDStr.(string) == userIDStr {
		// Further check if this is the only SuperAdmin or the primary one
		var user models.AdminUser
		config.DB.First(&user, "id = ?", userID)
		if user.Role == models.RoleSuperAdmin {
			var count int64
			config.DB.Model(&models.AdminUser{}).Where("role = ? AND deleted_at IS NULL", models.RoleSuperAdmin).Count(&count)
			if count <= 1 {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete the only SuperAdmin account"})
				return
			}
		}
	}

	result := config.DB.Delete(&models.AdminUser{}, "id = ?", userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete admin user: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found or already deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin user deleted successfully"})
}

