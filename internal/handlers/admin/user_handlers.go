package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
)

// CreateAdminUserRequest struct for creating a new admin user
type CreateAdminUserRequest struct {
	Email     string           `json:"email" binding:"required,email"`
	Password  string           `json:"password" binding:"required,min=8"` // Add more password complexity rules if needed
	FirstName string           `json:"firstName" binding:"required"`
	LastName  string           `json:"lastName" binding:"required"`
	Role      models.AdminUserRole `json:"role" binding:"required"`
}

// CreateAdminUser handles the creation of a new admin user (SuperAdmin only)
func CreateAdminUser(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Check if email already exists
	var existingUser models.AdminUser
	if config.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	hashedPassword, saltValue, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := models.AdminUser{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Salt:         saltValue,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		Status:       models.StatusActive, // Default to active, or require activation step
	}

	result := config.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user: " + result.Error.Error()})
		return
	}

	// Exclude password hash and salt from response
	newUserResponse := gin.H{
		"id":        newUser.ID,
		"email":     newUser.Email,
		"firstName": newUser.FirstName,
		"lastName":  newUser.LastName,
		"role":      newUser.Role,
		"status":    newUser.Status,
		"createdAt": newUser.CreatedAt,
		"updatedAt": newUser.UpdatedAt,
	}

	c.JSON(http.StatusCreated, newUserResponse)
}

// ListAdminUsers handles listing all admin users (SuperAdmin only)
func ListAdminUsers(c *gin.Context) {
	var users []models.AdminUser
	// Add pagination later if needed
	result := config.DB.Select("id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at").Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin users: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetAdminUser handles retrieving a single admin user by ID (SuperAdmin only)
func GetAdminUser(c *gin.Context) {
	userID := c.Param("id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user models.AdminUser
	result := config.DB.Select("id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at").Where("id = ?", parsedUserID).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateAdminUserRequest struct for updating an admin user
type UpdateAdminUserRequest struct {
	FirstName *string               `json:"firstName,omitempty"`
	LastName  *string               `json:"lastName,omitempty"`
	Role      *models.AdminUserRole `json:"role,omitempty"`
	// Password updates should be handled separately via a dedicated endpoint for security
}

// UpdateAdminUser handles updating an admin user (SuperAdmin only)
func UpdateAdminUser(c *gin.Context) {
	userID := c.Param("id")
	parsedUserID, err := uuid.Parse(userID)
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
	if config.DB.Where("id = ?", parsedUserID).First(&user).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

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

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided"})
		return
	}

	result := config.DB.Model(&user).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin user: " + result.Error.Error()})
		return
	}

	// Refetch to get updated data, excluding sensitive fields
	config.DB.Select("id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at").First(&user, "id = ?", parsedUserID)
	c.JSON(http.StatusOK, user)
}

// UpdateAdminUserStatusRequest struct for updating user status
type UpdateAdminUserStatusRequest struct {
	Status models.UserStatus `json:"status" binding:"required"` // Corrected type
}

// UpdateAdminUserStatus handles updating an admin user's status (SuperAdmin only)
func UpdateAdminUserStatus(c *gin.Context) {
	userID := c.Param("id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req UpdateAdminUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate status value
	if req.Status != models.StatusActive && 
	   req.Status != models.StatusInactive && 
	   req.Status != models.StatusLocked { // Corrected to StatusLocked
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	var user models.AdminUser
	if config.DB.Where("id = ?", parsedUserID).First(&user).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	// Prevent SuperAdmin from deactivating their own account if they are the only SuperAdmin (add this logic if needed)

	result := config.DB.Model(&user).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin user status: " + result.Error.Error()})
		return
	}

	// Refetch to get updated data, excluding sensitive fields
	config.DB.Select("id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at").First(&user, "id = ?", parsedUserID)
	c.JSON(http.StatusOK, user)
}

// DeleteAdminUser handles deleting an admin user (SuperAdmin only) - Soft delete is preferred
func DeleteAdminUser(c *gin.Context) {
	userID := c.Param("id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Prevent SuperAdmin from deleting their own account (especially if only one)
	// Add logic to check if the user being deleted is the one making the request
	// currentUserID := c.GetString("userID") // From JWTMiddleware
	// if currentUserID == userID { ... }

	result := config.DB.Delete(&models.AdminUser{}, "id = ?", parsedUserID)
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

