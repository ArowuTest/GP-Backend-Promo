package auth_test

import (
	"testing"
	"time"

	// "github.com/golang-jwt/jwt/v5" // This was unused as ValidateJWT is in the same package
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/auth"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateJWT(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := models.SuperAdminRole

	// jwtKey is initialized in auth.go init() function, relying on that for tests.
	// Ensure JWT_SECRET_KEY env var is not set or set to a known value if testing specific key behavior outside of default.

	user := &models.AdminUser{
		ID:    userID,
		Email: email,
		Role:  role,
	}

	tokenString, err := auth.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if tokenString == "" {
		t.Errorf("GenerateJWT() returned empty token string")
	}

	// Validate the token
	// We need to access the jwtKey used by the auth package or have a way to provide it for parsing.
	// For this test, we assume the auth package uses its initialized jwtKey.
	// If ValidateJWT is also in the auth package, it will use the same key.
	claims, err := auth.ValidateJWT(tokenString)

	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if claims.UserID != userID.String() { // Compare string representation of UUID
		t.Errorf("Expected UserID %v, got %v", userID.String(), claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}

	expectedExpiry := time.Now().Add(24 * time.Hour).Unix()
	if claims.ExpiresAt.Unix() > expectedExpiry+60 || claims.ExpiresAt.Unix() < expectedExpiry-60 { // Allow 1 min diff
		t.Errorf("Token expiry time is not approximately 24 hours from now. Got %v, Expected around %v", claims.ExpiresAt.Unix(), expectedExpiry)
	}
}

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "P@$$wOrd123"

	hashedPassword, salt, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hashedPassword == "" {
		t.Errorf("HashPassword() returned empty hashed password")
	}
	if salt == "" {
		t.Errorf("HashPassword() returned empty salt")
	}

	// Test correct password
	if !auth.CheckPasswordHash(password, salt, hashedPassword) { // Corrected function name and order of args
		t.Errorf("CheckPasswordHash() failed for correct password")
	}

	// Test incorrect password
	if auth.CheckPasswordHash("wrongpassword", salt, hashedPassword) { // Corrected function name and order of args
		t.Errorf("CheckPasswordHash() succeeded for incorrect password")
	}
}

func TestHashPassword_BcryptError(t *testing.T) {
    longPassword := make([]byte, 100) // bcrypt has a limit of 72 bytes for the password
    for i := range longPassword {
        longPassword[i] = 'a'
    }
    _, _, err := auth.HashPassword(string(longPassword))
    if err == nil {
        t.Errorf("HashPassword() did not error with a password longer than 72 bytes, expected bcrypt.ErrPasswordTooLong")
    } else if err != bcrypt.ErrPasswordTooLong {
         t.Errorf("HashPassword() with long password errored, but not with bcrypt.ErrPasswordTooLong. Error: %v", err)
    } else {
        // This is the expected outcome
        t.Logf("HashPassword() with long password errored with bcrypt.ErrPasswordTooLong as expected.")
    }
}

