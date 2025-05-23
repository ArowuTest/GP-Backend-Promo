package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims defines the claims for JWT tokens
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Role     string    `json:"role"`      // Single role for backward compatibility
	Roles    []string  `json:"roles"`     // Array of roles for future extensibility
	jwt.RegisteredClaims
}

// AuthMiddleware provides authentication middleware for the API
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	// If no secret is provided, use a default (but this should be avoided in production)
	if jwtSecret == "" {
		jwtSecret = "mynumba-donwin-jwt-secret-key-2025"
	}
	
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// Authenticate middleware checks if the request has a valid JWT token
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
		c.Set("username", claims.Username)
		
		// Store both role formats for maximum compatibility
		c.Set("userRole", claims.Role)
		c.Set("userRoles", claims.Roles)
		
		// Token is valid, continue
		c.Next()
	}
}

// RequireRole checks if the authenticated user has one of the required roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize an empty roles slice
		var userRoles []string
		
		// First try to get roles array from context
		rolesInterface, exists := c.Get("userRoles")
		if exists {
			// Try to convert to string slice
			roles, ok := rolesInterface.([]string)
			if ok && len(roles) > 0 {
				userRoles = roles
			}
		}
		
		// If userRoles is still empty, try to get single role
		if len(userRoles) == 0 {
			roleInterface, exists := c.Get("userRole")
			if exists {
				// Try to convert to string
				role, ok := roleInterface.(string)
				if ok && role != "" {
					userRoles = []string{role}
				}
			}
		}
		
		// If still no roles found, return unauthorized
		if len(userRoles) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false, 
				"error": "Unauthorized", 
				"details": "Authentication required or no role assigned",
			})
			c.Abort()
			return
		}
		
		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if contains(userRoles, requiredRole) {
				hasRequiredRole = true
				break
			}
		}
		
		if !hasRequiredRole {
			// Add debug information to help diagnose role issues
			c.JSON(http.StatusForbidden, gin.H{
				"success": false, 
				"error": "Forbidden", 
				"details": "Insufficient permissions",
				"user_roles": userRoles,
				"required_roles": requiredRoles,
			})
			c.Abort()
			return
		}
		
		// User has required role, continue
		c.Next()
	}
}

// GenerateJWT generates a new JWT token for the given user
func (m *AuthMiddleware) GenerateJWT(userID uuid.UUID, email string, username string, role string) (string, error) {
	// Create the claims with both single role and roles array for maximum compatibility
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Role:     role,
		Roles:    []string{role}, // Convert single role to array
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
	
	// Ensure backward compatibility - if Roles is empty but Role is set, populate Roles
	if len(claims.Roles) == 0 && claims.Role != "" {
		claims.Roles = []string{claims.Role}
	}
	
	// Ensure backward compatibility - if Role is empty but Roles has values, set Role to first value
	if claims.Role == "" && len(claims.Roles) > 0 {
		claims.Role = claims.Roles[0]
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
