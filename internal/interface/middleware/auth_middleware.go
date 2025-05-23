package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTClaims represents the claims in the JWT token
// MUST match the structure in authenticate_user.go
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Roles    []string  `json:"roles"`
	Role     string    `json:"role"`     // Added for backward compatibility with frontend
	Username string    `json:"username"` // Added for frontend use
	jwt.RegisteredClaims
}

// AuthMiddleware handles authentication for API requests
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	// If jwtSecret is empty, try to get it from environment
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// Fallback to default
			jwtSecret = "your-secret-key"
		}
	}
	
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// Authenticate checks if the request has a valid authentication token
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Parse and validate the JWT token
		claims, err := m.validateJWT(tokenString)
		if err != nil {
			// Enhanced error logging for debugging token issues
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false, 
				"error": "Invalid token", 
				"details": err.Error(),
				"message": "Your session has expired or is invalid. Please log in again.",
			})
			c.Abort()
			return
		}
		
		// Store user info in context for handlers to use
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRoles", claims.Roles)
		c.Set("userRole", claims.Role)       // Added for backward compatibility
		c.Set("username", claims.Username)   // Added for frontend use
		
		// Token is valid, continue
		c.Next()
	}
}

// RequireRole checks if the authenticated user has one of the required roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context (set by Authenticate middleware)
		rolesInterface, exists := c.Get("userRoles")
		if !exists {
			// Try to get single role for backward compatibility
			roleInterface, exists := c.Get("userRole")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized", "details": "Authentication required"})
				c.Abort()
				return
			}
			
			// Convert single role to roles array
			role, ok := roleInterface.(string)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Internal Server Error", "details": "Invalid role format"})
				c.Abort()
				return
			}
			
			rolesInterface = []string{role}
		}
		
		roles, ok := rolesInterface.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Internal Server Error", "details": "Invalid role format"})
			c.Abort()
			return
		}
		
		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if contains(roles, requiredRole) {
				hasRequiredRole = true
				break
			}
		}
		
		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Forbidden", "details": "Insufficient permissions"})
			c.Abort()
			return
		}
		
		// User has required role, continue
		c.Next()
	}
}

// GenerateJWT generates a new JWT token for the given user
func (m *AuthMiddleware) GenerateJWT(userID uuid.UUID, email string, roles []string) (string, error) {
	// Create the claims
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		// Note: Role and Username should be set when using this method directly
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mynumba-donwin-api",
			Subject:   userID.String(),
		},
	}
	
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(m.jwtSecret))
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// validateJWT validates the JWT token and returns the claims
func (m *AuthMiddleware) validateJWT(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		// Return the secret key
		return []byte(m.jwtSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Check if the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	// Extract the claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	
	// Verify expiration time
	if claims.ExpiresAt != nil {
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token has expired")
		}
	}
	
	return claims, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
