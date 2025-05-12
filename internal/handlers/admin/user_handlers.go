package admin

import (
	"net/http"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdminUser godoc
// @Summary Create a new admin user
// @Description Create a new admin user with username, password, email, first/last name, and role.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param user body models.AdminUser true "AdminUser object to be created. ID, CreatedAt, UpdatedAt, DeletedAt are ignored."
// @Success 201 {object} models.AdminUserResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [post]
func CreateAdminUser(c *gin.Context) {
	var req models.CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate role
	switch req.Role {
	case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
		// Valid role
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password: " + err.Error()})
		return
	}

	newUser := models.AdminUser{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		IsActive:  req.IsActive, // Default to true or based on request
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	responseUser := models.AdminUserResponse{
		ID:        newUser.ID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Role:      newUser.Role,
		IsActive:  newUser.IsActive,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	c.JSON(http.StatusCreated, responseUser)
}

// GetAdminUser godoc
// @Summary Get an admin user by ID
// @Description Get details of a specific admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.AdminUserResponse
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
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	responseUser := models.AdminUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	c.JSON(http.StatusOK, responseUser)
}

// UpdateAdminUser godoc
// @Summary Update an existing admin user
// @Description Update details of an existing admin user.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body models.UpdateAdminUserRequest true "AdminUser object with updated fields"
// @Success 200 {object} models.AdminUserResponse
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
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req models.UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Update fields if provided in the request
	if req.Username != nil {
		existingUser.Username = *req.Username
	}
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.FirstName != nil {
		existingUser.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existingUser.LastName = *req.LastName
	}
	if req.Role != nil {
		switch *req.Role {
		case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
			existingUser.Role = *req.Role
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
			return
		}
	}
	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	// Handle password update if provided
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password: " + err.Error()})
			return
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	responseUser := models.AdminUserResponse{
		ID:        existingUser.ID,
		Username:  existingUser.Username,
		Email:     existingUser.Email,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		Role:      existingUser.Role,
		IsActive:  existingUser.IsActive,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}
	c.JSON(http.StatusOK, responseUser)
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

	// Soft delete
	if err := config.DB.Delete(&models.AdminUser{}, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListAdminUsers godoc
// @Summary List all admin users
// @Description Get a list of all admin users.
// @Tags AdminUsers
// @Produce json
// @Success 200 {array} models.AdminUserResponse
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [get]
func ListAdminUsers(c *gin.Context) {
	var users []models.AdminUser
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
		return
	}

	responseUsers := make([]models.AdminUserResponse, len(users))
	for i, user := range users {
		responseUsers[i] = models.AdminUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, responseUsers)
}

// Login godoc
// @Summary Admin user login
// @Description Authenticate an admin user and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post]
func Login(c *gin.Context) {
	var creds models.LoginRequest

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var user models.AdminUser
	if err := config.DB.Where("username = ? AND is_active = ?", creds.Username, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username, password, or user is inactive"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateJWT(user.ID.String(), user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token, UserID: user.ID, Username: user.Username, Role: user.Role})
}

