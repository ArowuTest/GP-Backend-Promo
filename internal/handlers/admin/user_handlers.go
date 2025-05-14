package admin

import (
	"errors"
	"fmt" // Added for V4 logging
	"net/http"
	"os" // Added for V4 logging
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	// "golang.org/x/crypto/bcrypt" // Original was commented, V4 modifications will use auth.CheckPasswordHash
	"gorm.io/gorm"
)

// CreateAdminUserRequest defines the expected request body for creating an admin user
type CreateAdminUserRequest struct {
	Username  string                `json:"username" binding:"required"`
	Email     string                `json:"email" binding:"required,email"`
	Password  string                `json:"password" binding:"required,min=8"`
	FirstName string                `json:"first_name,omitempty"`
	LastName  string                `json:"last_name,omitempty"`
	Role      models.AdminUserRole  `json:"role" binding:"required"`
	Status    models.UserStatus     `json:"status,omitempty"` // e.g., Active, Inactive
}

// AdminUserResponse defines the structure for admin user responses, omitting sensitive data
type AdminUserResponse struct {
	ID          uuid.UUID             `json:"id"`
	Username    string                `json:"username"`
	Email       string                `json:"email"`
	FirstName   string                `json:"first_name,omitempty"`
	LastName    string                `json:"last_name,omitempty"`
	Role        models.AdminUserRole  `json:"role"`
	Status      models.UserStatus     `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	LastLoginAt *time.Time            `json:"last_login_at,omitempty"`
}

// toAdminUserResponse converts an AdminUser model to an AdminUserResponse
func toAdminUserResponse(user *models.AdminUser) AdminUserResponse {
	return AdminUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Role:        user.Role,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
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
		// Valid role
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	if req.Status == "" {
		req.Status = models.StatusActive // Default to Active if not provided
	} else {
		switch req.Status {
		case models.StatusActive, models.StatusInactive, models.StatusLocked:
			// Valid status
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
	Email     *string               `json:"email,omitempty"`      // Add email validation if allowing update
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
		hashedPassword, err := auth.HashPassword(*req.Password, existingUser.Salt) // Use existing salt
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password: " + err.Error()})
			return
		}
		updates["password_hash"] = hashedPassword
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now() // Explicitly set updated_at
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
// This struct will be used by the modified Login function.
// It matches the one from the V4 diagnostic, where Username can be username or email.
type LoginRequest struct {
	Username string `json:"username"` // Can be username or email
	Password string `json:"password" binding:"required"`
}

// LoginResponse defines the structure for login responses
// This is kept from the original file, though the V4 Login will return gin.H directly.
// If other functions use this, it should remain. For now, it is harmless.
type LoginResponse struct {
	Token    string               `json:"token"`
	UserID   uuid.UUID            `json:"user_id"`
	Username string               `json:"username"`
	Role     models.AdminUserRole `json:"role"`
}

// Login godoc
// @Summary Admin user login (V4 Diagnostic Query Test)
// @Description Authenticate an admin user and return a JWT token. Uses direct email lookup.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials (username field can be email)"
// @Success 200 {object} gin.H{"message": string, "token": string, "user": object}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 401 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/login [post] // This matches the original, assuming it's /api/v1/auth/login in router
func Login(c *gin.Context) {
	fmt.Fprintf(os.Stderr, "DEBUG: LOGIN HANDLER ENTERED (V4 Query Test)\n")
	var req LoginRequest // Uses the LoginRequest defined above in this file

	// Bind JSON payload to LoginRequest struct
	if err := c.ShouldBindJSON(&req); err != nil {
		// Attempt to read and log raw body for debugging, then reset for Gin
		// Note: Reading c.Request.Body directly consumes it. This is for debug only.
		// Consider using c.GetRawData() and then c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// For simplicity in diagnostic, just logging the error is primary.
		fmt.Fprintf(os.Stderr, "DEBUG: Raw Login Request Body access attempted on Binding Error (V4 Query Test). Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "LOGIN PAYLOAD BINDING ERROR (V4 Query Test): " + err.Error()})
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Login Request Payload (V4 Query Test): Username/Email=\"%s\", Password Present=%t\n", req.Username, req.Password != "")

	// Ensure Username (which holds email) is provided
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LOGIN PAYLOAD ERROR (V4 Query Test): Username or Email is required"})
		return
	}

	var user models.AdminUser
	loginID := req.Username // This will contain the email address from the frontend

	fmt.Fprintf(os.Stderr, "DEBUG: Attempting DB lookup with loginID (V4 Query Test): 	'%s'\n", loginID)

	// V4 DIAGNOSTIC: Simplified query - directly look for email and active status
	if err := config.DB.Where("email = ? AND status = ?", loginID, models.StatusActive).First(&user).Error; err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Database lookup error (V4 Query Test) for loginID '%s': %v\n", loginID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username, password, or user is not active (V4 Query Test - DB Error)"})
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: User found in DB (V4 Query Test): UserID=%s, Email=%s, Status=%s\n", user.ID.String(), user.Email, user.Status)

	// Compare the provided password with the stored hash using the original auth.CheckPasswordHash
	if !auth.CheckPasswordHash(req.Password, user.Salt, user.PasswordHash) {
		fmt.Fprintf(os.Stderr, "DEBUG: Password comparison failed (V4 Query Test) for UserID '%s' using auth.CheckPasswordHash\n", user.ID.String())
		// Original logic for failed login attempts
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 { // Assuming 5 is the threshold
			user.Status = models.StatusLocked
		}
		config.DB.Save(&user)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username, password, or user is not active (V4 Query Test - Pwd Fail)"})
		return
	}

	// Reset failed login attempts and update LastLoginAt as in original
	user.FailedLoginAttempts = 0
	now := time.Now()
    user.LastLoginAt = &now 
	config.DB.Save(&user)

	// Generate JWT token using original auth.GenerateJWT
	// Assuming auth.GenerateJWT takes (userID string, username string, role models.AdminUserRole)
	token, err := auth.GenerateJWT(user.ID.String(), user.Username, user.Role) 
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: JWT generation failed (V4 Query Test) for UserID '%s': %v\n", user.ID.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token (V4 Query Test)"})
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Login successful, token generated (V4 Query Test) for UserID '%s'\n", user.ID.String())
	// Return gin.H as per V4 diagnostic, not LoginResponse struct, for this test
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful (V4 Query Test)",
		"token":   token,
		"user": gin.H{
			"id":        user.ID.String(), // Ensure ID is string if gin.H expects it
			"email":     user.Email,
			"username":  user.Username, // Added username to response
			"role":      user.Role,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		},
	})
}

