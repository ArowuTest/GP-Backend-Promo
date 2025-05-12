package handlers

import (
	"net/http"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUser godoc
// @Summary Create a new admin user
// @Description Create a new admin user with username, password, email, and role.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param user body models.AdminUser true "AdminUser object to be created. Password field is for input only."
// @Success 201 {object} models.AdminUser "User created successfully (PasswordHash and Salt excluded)"
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [post]
func CreateUser(c *gin.Context) {
	var newUser models.AdminUser
	// Bind only specific fields for creation to avoid unexpected inputs
	var input struct {
		Username string             `json:"username" binding:"required"`
		Email    string             `json:"email" binding:"required,email"`
		Password string             `json:"password" binding:"required,min=6"`
		Role     models.AdminUserRole `json:"role" binding:"required"`
		IsActive bool               `json:"isActive"` // Default will be handled by model or DB
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate role
	switch input.Role {
	case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
		// Valid role
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password: " + err.Error()})
		return
	}

	newUser.Username = input.Username
	newUser.Email = input.Email
	newUser.PasswordHash = string(hashedPassword)
	newUser.Role = input.Role
	// newUser.Status will default to 'Active' as per model; Salt is not explicitly set as bcrypt includes it in the hash
	// newUser.IsActive from input can be mapped to newUser.Status if needed, e.g.
	if input.IsActive {
	    newUser.Status = models.StatusActive
	} else {
	    newUser.Status = models.StatusInactive // Or some other default for new users if not active
	}


	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// PasswordHash and Salt are excluded from JSON response due to `json:"-"` tags in the model
	c.JSON(http.StatusCreated, newUser)
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get details of a specific admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.AdminUser "User details (PasswordHash and Salt excluded)"
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Router /admin/users/{id} [get]
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format. Expected UUID."})
		return
	}

	var user models.AdminUser
	if err := config.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
		    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Update details of an existing admin user (username, email, role, status). Password update should be a separate endpoint.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body models.AdminUser true "AdminUser object with updated fields (username, email, role, status)"
// @Success 200 {object} models.AdminUser "User updated successfully (PasswordHash and Salt excluded)"
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format. Expected UUID."})
		return
	}

	var existingUser models.AdminUser
	if err := config.DB.First(&existingUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
		    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user for update: " + err.Error()})
		}
		return
	}

	var updatedInfo struct {
		Username  string             `json:"username"`
		Email     string             `json:"email"`
		Role      models.AdminUserRole `json:"role"`
		Status    models.UserStatus  `json:"status"`
        Password  string             `json:"password,omitempty"` // For password change
	}

	if err := c.ShouldBindJSON(&updatedInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Update fields if provided
	if updatedInfo.Username != "" {
		existingUser.Username = updatedInfo.Username
	}
	if updatedInfo.Email != "" {
		existingUser.Email = updatedInfo.Email
	}
	if updatedInfo.Role != "" {
		switch updatedInfo.Role {
		case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
			existingUser.Role = updatedInfo.Role
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
			return
		}
	}
	if updatedInfo.Status != "" {
	    switch updatedInfo.Status {
	    case models.StatusActive, models.StatusInactive, models.StatusLocked:
	        existingUser.Status = updatedInfo.Status
	    default:
	        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user status specified"})
			return
	    }
	}

    if updatedInfo.Password != "" {
        if len(updatedInfo.Password) < 6 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters long"})
            return
        }
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedInfo.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password: " + err.Error()})
            return
        }
        existingUser.PasswordHash = string(hashedPassword)
    }

	if err := config.DB.Save(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, existingUser)
}

// DeleteUser godoc
// @Summary Delete a user by ID
// @Description Delete an admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} gin.H{"message": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string} "User not found or already deleted"
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format. Expected UUID."})
		return
	}

    // Check if user exists before attempting delete to provide a 404 if not found
    var user models.AdminUser
    if err := config.DB.First(&user, userID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence: " + err.Error()})
        }
        return
    }

	// GORM's Delete with a struct pointer will use primary key for deletion.
	// For soft delete, ensure DeletedAt field is in your model and GORM is configured for it.
	if err := config.DB.Delete(&models.AdminUser{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers godoc
// @Summary List all admin users
// @Description Get a list of all admin users.
// @Tags AdminUsers
// @Produce json
// @Success 200 {array} models.AdminUser "List of users (PasswordHash and Salt excluded)"
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [get]
func ListUsers(c *gin.Context) {
	var users []models.AdminUser
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body struct{Username string `json:"username" binding:"required"`; Password string `json:"password" binding:"required"`} true "Login credentials"
// @Success 200 {object} gin.H{"token": string, "user_id": string, "role": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post]
func Login(c *gin.Context) {
	var creds struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var user models.AdminUser
	if err := config.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
		    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login: " + err.Error()})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateJWT(user.ID.String(), user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID.String(), "role": user.Role})
}

