package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

func init() {
	// In a real application, get this from environment variables
	// For now, using a placeholder. This MUST be changed for production.
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		fmt.Println("Warning: JWT_SECRET_KEY not set, using default insecure key. SET THIS FOR PRODUCTION!")
		secret = "your-very-secret-and-long-key-that-is-at-least-32-bytes"
	}
	jwtKey = []byte(secret)
}

// Claims struct for JWT
type Claims struct {
	UserID string           `json:"user_id"`
	Email  string           `json:"email"`
	Role   models.AdminUserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a given user
func GenerateJWT(user *models.AdminUser) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateJWT validates a JWT token string
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// JWTMiddleware provides JWT authentication for Gin routes
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := ValidateJWT(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Set user context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// HashPassword hashes a given password using bcrypt
func HashPassword(password string) (string, string, error) {
	// Generate a salt with default cost
	saltBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Using password as salt source for simplicity here, usually a separate random salt
    if err != nil {
        return "", "", err
    }
    salt := string(saltBytes[:16]) // Example: take first 16 bytes as salt string, can be more robust

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return string(hashedPassword), salt, nil
}

// CheckPasswordHash checks if the provided password matches the stored hash and salt
func CheckPasswordHash(password, salt, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password))
	return err == nil
}

// LoginRequest struct for login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginAdminUser handles admin user login
func LoginAdminUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var adminUser models.AdminUser
	result := config.DB.Where("email = ?", req.Email).First(&adminUser)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if adminUser.Status != models.StatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
		return
	}

	if !CheckPasswordHash(req.Password, adminUser.Salt, adminUser.PasswordHash) {
		// Implement failed login attempt tracking here if needed
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := GenerateJWT(&adminUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update LastLoginAt
	now := time.Now()
	adminUser.LastLoginAt = &now
	adminUser.FailedLoginAttempts = 0 // Reset on successful login
	config.DB.Save(&adminUser)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":        adminUser.ID,
			"email":     adminUser.Email,
			"role":      adminUser.Role,
			"firstName": adminUser.FirstName,
			"lastName":  adminUser.LastName,
		},
	})
}

// Placeholder for RegisterAdminUser - to be implemented if self-registration is allowed or for seeding
// func RegisterAdminUser(c *gin.Context) { ... }




// RoleAuthMiddleware checks if the authenticated user has one of the required roles
func RoleAuthMiddleware(requiredRoles ...models.AdminUserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleContext, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found in context. Ensure JWTMiddleware runs first."})
			c.Abort()
			return
		}

		currentRole, ok := userRoleContext.(models.AdminUserRole)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User role in context is of an unexpected type"})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range requiredRoles {
			if currentRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}

