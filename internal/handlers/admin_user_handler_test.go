package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	// "time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ArowuTest/GP-Backend-Promo/internal/auth" // For JWT and roles
	"github.com/ArowuTest/GP-Backend-Promo/internal/config" // For DB (mocked or test instance)
	"github.com/ArowuTest/GP-Backend-Promo/internal/handlers" // Handlers to test
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	// "github.com/google/uuid"
)

// Helper function to set up a Gin router for testing
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// config.ConnectDB() // For actual DB, for tests we might mock this
	// For now, let's assume DB is handled or mocked elsewhere or tests are designed to not hit DB directly without setup
	return router
}

// Helper to create a request with a JWT token for a specific role
func createAuthenticatedRequest(t *testing.T, method, path string, body []byte, role models.UserRole) *http.Request {
	// Generate a token for the given role
	// For testing, the username in the token can be a dummy one
	token, err := auth.GenerateJWT("testuser", role)
	assert.NoError(t, err)

	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// TestCreateUser - Example test for CreateUser handler
func TestCreateUser_SuperAdmin(t *testing.T) {
    // Setup: Initialize a test database or mock the DB interactions
    // For this example, we assume a mock DB or a test DB is configured via config.ConnectDB() if it were called.
    // If using a real test DB, ensure it's clean before each test or use transactions.
    config.ConnectDB() // This will use your actual DB connection logic. Be careful with test data.
    // Ideally, use a dedicated test database configured via environment variables.

	router := setupTestRouter()
	adminApi := router.Group("/admin")
    adminApi.Use(auth.JWTMiddleware()) // Apply JWT middleware
    userManagement := adminApi.Group("/users")
    userManagement.Use(auth.RoleAuthMiddleware(models.SuperAdminRole))
    {
        userManagement.POST("/", handlers.CreateUser)
    }

	newUser := models.AdminUserInput{
		Username:  "testnewuser",
		Email:     "testnewuser@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "UserNew",
		Role:      models.AdminRole, // SuperAdmin creating an Admin
	}
	jsonBody, _ := json.Marshal(newUser)

	req := createAuthenticatedRequest(t, "POST", "/admin/users/", jsonBody, models.SuperAdminRole)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdUserResponse struct {
        Message string             `json:"message"`
        User    models.AdminUser `json:"user"`
    }
	json.Unmarshal(w.Body.Bytes(), &createdUserResponse)
	assert.Equal(t, "Admin user created successfully", createdUserResponse.Message)
	assert.Equal(t, newUser.Email, createdUserResponse.User.Email)
    assert.Equal(t, newUser.Role, createdUserResponse.User.Role)

    // Cleanup: Delete the created user from the test database
    // config.DB.Unscoped().Delete(&models.AdminUser{}, "email = ?", newUser.Email)
}

func TestCreateUser_ForbiddenForAdminRole(t *testing.T) {
    config.ConnectDB()
	router := setupTestRouter()
	adminApi := router.Group("/admin")
    adminApi.Use(auth.JWTMiddleware())
    userManagement := adminApi.Group("/users")
    userManagement.Use(auth.RoleAuthMiddleware(models.SuperAdminRole))
    {
        userManagement.POST("/", handlers.CreateUser)
    }

	newUser := models.AdminUserInput{
		Username:  "testforbiddenuser",
		Email:     "testforbiddenuser@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "Forbidden",
		Role:      models.AdminRole,
	}
	jsonBody, _ := json.Marshal(newUser)

	req := createAuthenticatedRequest(t, "POST", "/admin/users/", jsonBody, models.AdminRole) // Attempting as Admin
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Add more tests for ListUsers, GetUser, UpdateUser, DeleteUser with different roles and scenarios
// For example:
// - TestListUsers_SuperAdmin_Success
// - TestGetUser_SuperAdmin_Success
// - TestGetUser_NotFound
// - TestUpdateUser_SuperAdmin_Success
// - TestUpdateUser_ForbiddenForAdmin
// - TestDeleteUser_SuperAdmin_Success
// - TestDeleteUser_ForbiddenForAdmin

// Remember to handle database state carefully. Using a separate test database is highly recommended.
// Mocking the database layer (GORM) is another common approach for unit tests to avoid DB dependency.

