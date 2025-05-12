package handlers

import (
	"net/http"
	"strconv"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser godoc
// @Summary Create a new admin user
// @Description Create a new admin user with username, password, and role.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param user body models.User true "User object to be created"
// @Success 201 {object} models.User
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [post]
func CreateUser(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate role
	switch newUser.Role {
	case models.SuperAdminRole, models.AdminRole, models.SeniorUserRole, models.WinnerReportsUserRole, models.AllReportUserRole:
		// Valid role
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password: " + err.Error()})
		return
	}
	newUser.Password = string(hashedPassword)

	if err := db.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Clear password before returning
	newUser.Password = ""
	c.JSON(http.StatusCreated, newUser)
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get details of a specific admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Router /admin/users/{id} [get]
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Clear password before returning
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Update details of an existing admin user (username, role). Password update should be a separate endpoint.
// @Tags AdminUsers
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.User true "User object with updated fields (username, role)"
// @Success 200 {object} models.User
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var existingUser models.User
	if err := db.DB.First(&existingUser, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updatedInfo models.User
	if err := c.ShouldBindJSON(&updatedInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Update fields if provided
	if updatedInfo.Username != "" {
		existingUser.Username = updatedInfo.Username
	}
	if updatedInfo.Role != "" {
		// Validate role
		switch updatedInfo.Role {
		case models.SuperAdminRole, models.AdminRole, models.SeniorUserRole, models.WinnerReportsUserRole, models.AllReportUserRole:
			existingUser.Role = updatedInfo.Role
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role specified"})
			return
		}
	}

	if err := db.DB.Save(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	// Clear password before returning
	existingUser.Password = ""
	c.JSON(http.StatusOK, existingUser)
}

// DeleteUser godoc
// @Summary Delete a user by ID
// @Description Delete an admin user by their ID.
// @Tags AdminUsers
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} gin.H{"message": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	if err := db.DB.Delete(&models.User{}, uint(id)).Error; err != nil {
		// Check if the error is because the record was not found, or some other DB error
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
// @Success 200 {array} models.User
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/users [get]
func ListUsers(c *gin.Context) {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
		return
	}

	// Clear passwords before returning
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, users)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body struct{Username string `json:"username"`; Password string `json:"password"`} true "Login credentials"
// @Success 200 {object} gin.H{"token": string}
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

	var user models.User
	if err := db.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateJWT(user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

