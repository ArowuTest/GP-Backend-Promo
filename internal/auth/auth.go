package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models" // Use the central models package
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func init() {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		fmt.Println("Warning: JWT_SECRET_KEY not set, using default insecure key. SET THIS FOR PRODUCTION!")
		secret = "a-very-secure-secret-key-for-jwt-must-be-long-enough"
	}
	jwtKey = []byte(secret)
}

// Claims struct for JWT
type Claims struct {
	Username string                 `json:"username"`
	Role     models.AdminUserRole   `json:"role"` // Corrected to AdminUserRole
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a given username and role
func GenerateJWT(username string, role models.AdminUserRole) (string, error) { // Corrected to AdminUserRole
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
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
		c.Set("username", claims.Username) // Set username from claims
		c.Set("userRole", claims.Role)     // Set role from claims

		c.Next()
	}
}

// RoleAuthMiddleware checks if the authenticated user has one of the required roles
func RoleAuthMiddleware(requiredRoles ...models.AdminUserRole) gin.HandlerFunc { // Corrected to AdminUserRole
	return func(c *gin.Context) {
		userRoleContext, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found in context. Ensure JWTMiddleware runs first."})
			c.Abort()
			return
		}

		currentRole, ok := userRoleContext.(models.AdminUserRole) // Corrected to AdminUserRole
		if !ok {
			// Attempt to convert if it's a string (e.g., from admin_user_handler Login)
			currentRoleStr, okStr := userRoleContext.(string)
			if okStr {
				currentRole = models.AdminUserRole(currentRoleStr) // Corrected to AdminUserRole
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User role in context is of an unexpected type"})
				c.Abort()
				return
			}
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

