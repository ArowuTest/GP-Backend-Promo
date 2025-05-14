package handlers

import (
	"bytes"
	"errors" // Added for errors.Is
	"fmt"    // Added for fmt.Fprintf
	"io"
	// "log" // Replaced with fmt.Fprintf(os.Stderr, ...)
	"net/http"
	"os"     // Added for os.Stderr

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
// @Param credentials body struct{Username string `json:"username"`; Email string `json:"email"`; Password string `json:"password" binding:"required"`} true "Login credentials (username or email required)"
// @Success 200 {object} gin.H{"token": string, "user_id": string, "username": string, "role": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post] // Ensure this matches the frontend API call path /api/v1/auth/login
func Login(c *gin.Context) {
	fmt.Fprintf(os.Stderr, "DEBUG: LOGIN HANDLER ENTERED (v2 Error Test with Payload Fix)\n") // More reliable logging & version marker

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Error reading request body (v2 Error Test with Payload Fix): %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "DEBUG: Raw Login Request Body (v2 Error Test with Payload Fix): %s\n", string(bodyBytes))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var creds struct {
		Username string `json:"username"` // Username is now optional at binding stage
		Email    string `json:"email"`    // Email is now accepted
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Error binding JSON for login (v2 Error Test with Payload Fix): %v. Payload was: %s\n", err, string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "LOGIN PAYLOAD BINDING ERROR (v2 Payload Fix): " + err.Error()})
		return
	}

	if creds.Username == "" && creds.Email == "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Username and Email both empty (v2 Error Test with Payload Fix). Payload was: %s\n", string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "LOGIN PAYLOAD ERROR (v2 Payload Fix): Username or Email is required"})
		return
	}

	loginIdentifier := creds.Username
	if loginIdentifier == "" {
		loginIdentifier = creds.Email
	}
	fmt.Fprintf(os.Stderr, "DEBUG: Login credentials bound (v2 Error Test with Payload Fix): Identifier=\"%s\"\n", loginIdentifier)

	var user models.AdminUser
	// Query by username or email
	if err := config.DB.Where("username = ? OR email = ?", loginIdentifier, loginIdentifier).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Fprintf(os.Stderr, "DEBUG: User not found for identifier (v2 Error Test with Payload Fix): %s\n", loginIdentifier)
		    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"}) // More generic error
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Database error looking up user (v2 Error Test with Payload Fix) %s: %v\n", loginIdentifier, err)
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login: " + err.Error()})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Password mismatch for user (v2 Error Test with Payload Fix): %s\n", loginIdentifier)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"}) // More generic error
		return
	}
	fmt.Fprintf(os.Stderr, "DEBUG: User %s authenticated successfully (v2 Error Test with Payload Fix)\n", loginIdentifier)

	token, err := auth.GenerateJWT(user.ID.String(), user.Username, user.Role) // Use user.Username for JWT claim
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Error generating JWT for user (v2 Error Test with Payload Fix) %s: %v\n", loginIdentifier, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: JWT generated successfully for user (v2 Error Test with Payload Fix) %s\n", loginIdentifier)
	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID.String(), "username": user.Username, "role": user.Role})
	fmt.Fprintf(os.Stderr, "DEBUG: Login handler finished (v2 Error Test with Payload Fix)\n")
}

