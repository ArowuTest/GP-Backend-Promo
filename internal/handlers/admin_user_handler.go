package handlers

import (
	"bytes"
	"errors" // Added for errors.Is
	"io"
	"log"
	"net/http"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt" // Keep bcrypt as it's used for CompareHashAndPassword and GenerateFromPassword
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
	var input struct {
		Username string             `json:"username" binding:"required"`
		Email    string             `json:"email" binding:"required,email"`
		Password string             `json:"password" binding:"required,min=6"`
		Role     models.AdminUserRole `json:"role" binding:"required"`
		Status   models.UserStatus  `json:"status,omitempty"` // Allow status to be set, default to Active if not provided
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	switch input.Role {
	case models.RoleSuperAdmin, models.RoleAdmin, models.RoleSeniorUser, models.RoleWinnerReportsUser, models.RoleAllReportUser:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password: " + err.Error()})
		return
	}

	newUser.Username = input.Username
	newUser.Email = input.Email
	newUser.PasswordHash = string(hashedPassword)
	newUser.Role = input.Role
    if input.Status != "" {
        switch input.Status {
        case models.StatusActive, models.StatusInactive, models.StatusLocked:
            newUser.Status = input.Status
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user status specified"})
            return
        }
    } else {
        newUser.Status = models.StatusActive // Default to Active
    }

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}
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
		if errors.Is(err, gorm.ErrRecordNotFound) { // Use errors.Is
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
// @Description Update details of an existing admin user (username, email, role, status). Password update can also be done.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body object{username=string,email=string,role=models.AdminUserRole,status=models.UserStatus,password=string} true "AdminUser object with updated fields"
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
		if errors.Is(err, gorm.ErrRecordNotFound) { // Use errors.Is
		    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user for update: " + err.Error()})
		}
		return
	}

	var updatedInfo struct {
		Username  string             `json:"username,omitempty"`
		Email     string             `json:"email,omitempty"`
		Role      models.AdminUserRole `json:"role,omitempty"`
		Status    models.UserStatus  `json:"status,omitempty"`
        Password  string             `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&updatedInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

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

    var user models.AdminUser
    if err := config.DB.First(&user, userID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { // Use errors.Is
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence: " + err.Error()})
        }
        return
    }

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
// @Success 200 {object} gin.H{"token": string, "user_id": string, "username": string, "role": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post]
func Login(c *gin.Context) {
	log.Println("DEBUG: Login handler started")

	// Log the raw request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("DEBUG: Error reading request body: %v\n", err)
		// Don't return here, try to proceed if possible, or handle as appropriate
	} else {
		log.Printf("DEBUG: Raw Login Request Body: %s\n", string(bodyBytes))
		// After reading, we need to replace the body so Gin can read it again for binding
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var creds struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		log.Printf("DEBUG: Error binding JSON for login: %v. Payload was: %s\n", err, string(bodyBytes)) // Log body again on error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}
	log.Printf("DEBUG: Login credentials bound successfully: Username=\"%s\"", creds.Username)

	var user models.AdminUser
	if err := config.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DEBUG: User not found for username: %s\n", creds.Username)
		    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			log.Printf("DEBUG: Database error looking up user %s: %v\n", creds.Username, err)
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login: " + err.Error()})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		log.Printf("DEBUG: Password mismatch for user: %s\n", creds.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	log.Printf("DEBUG: User %s authenticated successfully\n", creds.Username)

	token, err := auth.GenerateJWT(user.ID.String(), user.Username, user.Role)
	if err != nil {
		log.Printf("DEBUG: Error generating JWT for user %s: %v\n", creds.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	log.Printf("DEBUG: JWT generated successfully for user %s\n", creds.Username)
	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID.String(), "username": user.Username, "role": user.Role})
	log.Println("DEBUG: Login handler finished")
}

