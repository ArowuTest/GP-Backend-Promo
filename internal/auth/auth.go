package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

const saltSize = 16

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
	UserID   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Role     models.AdminUserRole   `json:"role"`
	jwt.RegisteredClaims
}

// GenerateSalt creates a new random salt
func GenerateSalt() (string, error) {
	saltBytes := make([]byte, saltSize)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(saltBytes), nil
}

// HashPassword generates a bcrypt hash of the password using the salt.
// Note: bcrypt incorporates its own salt, so the provided salt here is more for a pepper or if a specific pre-generation salting step is desired.
// Standard bcrypt.GenerateFromPassword handles salting internally.
// For simplicity and standard bcrypt usage, we will not use the separate salt parameter directly in bcrypt.GenerateFromPassword
// but ensure the AdminUser model has a PasswordHash field.
// If the salt field in AdminUser model is intended to be combined *before* bcrypt, that logic would be different.
// Assuming standard bcrypt usage where it generates and embeds its own salt in the hash string.
// The `salt` parameter here will be ignored for standard bcrypt, but kept if the design intends a separate salt field in DB.
func HashPassword(password string, salt string) (string, error) { // Salt parameter might be vestigial if using bcrypt correctly
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a password with a bcrypt hash.
// The `salt` parameter is not used by bcrypt.CompareHashAndPassword as the salt is part of the hash itself.
func CheckPasswordHash(password string, salt string, hash string) bool { // Salt parameter is vestigial for bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a new JWT token for a given userID, username and role
func GenerateJWT(userID string, username string, role models.AdminUserRole) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID, // Using userID as subject
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

		c.Set("userID", claims.UserID) // Set userID from claims
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

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
			currentRoleStr, okStr := userRoleContext.(string)
			if okStr {
				currentRole = models.AdminUserRole(currentRoleStr)
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

