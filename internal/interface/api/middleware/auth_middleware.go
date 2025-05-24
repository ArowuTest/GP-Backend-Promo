package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Authenticate validates JWT token and sets user information in context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "Invalid authorization format, expected 'Bearer {token}'",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "Invalid token",
			})
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "Token has expired",
			})
			c.Abort()
			return
		}

		// Set user information in context
		// Convert userID to UUID for consistent handling
		userUUID, err := uuid.Parse(claims.UserID)
		if err == nil {
			c.Set("userID", userUUID)
		} else {
			// Fallback to string if parsing fails
			c.Set("userID", claims.UserID)
		}
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole checks if the user has the required role
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		roleInterface, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Details: "User role not found in token",
			})
			c.Abort()
			return
		}

		role, ok := roleInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Internal server error",
				Details: "Failed to parse user role",
			})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, r := range roles {
			if role == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, response.ErrorResponse{
				Success: false,
				Error:   "Forbidden",
				Details: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateToken generates a JWT token
func (m *AuthMiddleware) GenerateToken(userID, username, role string, expirationHours int) (string, time.Time, error) {
	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	// Create claims
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(m.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}
