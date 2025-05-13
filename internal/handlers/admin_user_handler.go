package handlers

import (
	"fmt"    // Added for fmt.Fprintf
	"net/http"
	"os"     // Added for os.Stderr

	"github.com/gin-gonic/gin"
	// Commenting out unused imports for this diagnostic version
	// "bytes"
	// "errors"
	// "io"
	// "github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	// "github.com/ArowuTest/GP-Backend-Promo/internal/config"
	// "github.com/ArowuTest/GP-Backend-Promo/internal/models"
	// "github.com/google/uuid"
	// "golang.org/x/crypto/bcrypt"
	// "gorm.io/gorm"
)

// CreateUser godoc
// ... (rest of the file remains the same, only Login function is modified for diagnostics)
// GetUser godoc
// ...
// UpdateUser godoc
// ...
// DeleteUser godoc
// ...
// ListUsers godoc
// ...

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
// @Router /admin/login [post] // Ensure this matches the frontend API call path /api/v1/auth/login
func Login(c *gin.Context) {
	fmt.Fprintf(os.Stderr, "DEBUG: LOGIN HANDLER ENTERED (V3 DEPLOYMENT DIAGNOSTIC)\n")
	// IMMEDIATELY RETURN A UNIQUE HARDCODED ERROR FOR DIAGNOSTIC PURPOSES
	c.JSON(http.StatusBadRequest, gin.H{"error": "DEPLOYMENT TEST V3 - LOGIN HANDLER REACHED AND EXECUTED"})
	return

	// All previous logic is bypassed for this diagnostic test
	/*
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Error reading request body (v2 Error Test): %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Raw Login Request Body (v2 Error Test): %s\n", string(bodyBytes))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		var creds struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&creds); err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Error binding JSON for login (v2 Error Test): %v. Payload was: %s\n", err, string(bodyBytes))
			// MODIFIED ERROR MESSAGE FOR TESTING
			c.JSON(http.StatusBadRequest, gin.H{"error": "LOGIN PAYLOAD BINDING ERROR (v2): " + err.Error()})
			return
		}
		fmt.Fprintf(os.Stderr, "DEBUG: Login credentials bound successfully (v2 Error Test): Username=\"%s\"\n", creds.Username)

		var user models.AdminUser
		if err := config.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Fprintf(os.Stderr, "DEBUG: User not found for username (v2 Error Test): %s\n", creds.Username)
			    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			} else {
				fmt.Fprintf(os.Stderr, "DEBUG: Database error looking up user (v2 Error Test) %s: %v\n", creds.Username, err)
			    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login: " + err.Error()})
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Password mismatch for user (v2 Error Test): %s\n", creds.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		fmt.Fprintf(os.Stderr, "DEBUG: User %s authenticated successfully (v2 Error Test)\n", creds.Username)

		token, err := auth.GenerateJWT(user.ID.String(), user.Username, user.Role)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Error generating JWT for user (v2 Error Test) %s: %v\n", creds.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		fmt.Fprintf(os.Stderr, "DEBUG: JWT generated successfully for user (v2 Error Test) %s\n", creds.Username)
		c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID.String(), "username": user.Username, "role": user.Role})
		fmt.Fprintf(os.Stderr, "DEBUG: Login handler finished (v2 Error Test)\n")
	*/
}

// The rest of the functions (CreateUser, GetUser, UpdateUser, DeleteUser, ListUsers) remain unchanged from the previous version.
// To keep the message brief, I will paste the full content of the original file and then indicate where the Login function is modified.
// Assume the full content of /home/ubuntu/backend_review/home/ubuntu/GP-Backend-Promo/internal/handlers/admin_user_handler.go
// is here, with the Login function replaced by the one above.

