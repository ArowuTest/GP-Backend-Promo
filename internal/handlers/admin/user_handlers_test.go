package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mynumba-don-win-draw-system/backend/internal/auth"
	"mynumba-don-win-draw-system/backend/internal/config"
	"mynumba-don-win-draw-system/backend/internal/models"
)

// Helper function to setup GORM with sqlmock
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
		PreferSimpleProtocol: true, // For sqlmock
	}), &gorm.Config{}) 
	assert.NoError(t, err)

	return mock, gormDB
}


func TestCreateAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, gormDB := setupMockDB(t)
	originalDB := config.DB // Backup original DB
	config.DB = gormDB      // Set mocked DB for the test
	defer func() {
		config.DB = originalDB // Restore original DB
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	// Mock the check for existing user (should not find one)
	mock.ExpectQuery(`SELECT \* FROM "admin_users" WHERE email = \$1 ORDER BY "admin_users"."id" LIMIT \$2`).
		WithArgs("testuser@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock the transaction for creating the user
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "admin_users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit() // Match the transaction commit

	router := gin.Default()
	// Assuming a middleware for JWT auth is applied at a higher level or within the handler group
	// For this test, we will manually create a token and pass it.
	// The CreateAdminUser handler itself might check for SuperAdmin role from context.
	router.POST("/admin/users", CreateAdminUser) 

	// Prepare the request body - this should match what CreateAdminUser expects
	// Assuming CreateAdminUser binds to a struct like this for the request:
	createUserReq := struct {
		Email     string             `json:"email" binding:"required,email"`
		Password  string             `json:"password" binding:"required,min=8"`
		Role      models.AdminUserRole `json:"role" binding:"required"`
		FirstName string             `json:"firstName" binding:"required"`
		LastName  string             `json:"lastName" binding:"required"`
	}{
		Email:     "testuser@example.com",
		Password:  "password123",
		Role:      models.DrawAdminRole,
		FirstName: "Test",
		LastName:  "User",
	}
	jsonValue, _ := json.Marshal(createUserReq)

	req, _ := http.NewRequest(http.MethodPost, "/admin/users", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// Create a SuperAdmin token for authorization
	// auth.GenerateJWT now expects *models.AdminUser
	dummySuperAdmin := &models.AdminUser{
		ID:    uuid.New(),
		Email: "super@example.com",
		Role:  models.SuperAdminRole, // Use the defined constant
	}
	superAdminToken, err := auth.GenerateJWT(dummySuperAdmin)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+superAdminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListAdminUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, gormDB := setupMockDB(t)
	originalDB := config.DB
	config.DB = gormDB
	defer func() {
		config.DB = originalDB
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "role", "status", "created_at", "updated_at"}).
		AddRow(uuid.New(), "user1@example.com", "DRAW_ADMIN", "ACTIVE", now, now).
		AddRow(uuid.New(), "user2@example.com", "SUPER_ADMIN", "ACTIVE", now, now)
	mock.ExpectQuery(`SELECT id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at FROM "admin_users"`).WillReturnRows(rows)
	router := gin.Default()
	router.GET("/admin/users", ListAdminUsers)

	// Create a SuperAdmin token for authorization
	dummySuperAdmin := &models.AdminUser{
		ID:    uuid.New(),
		Email: "super@example.com",
		Role:  models.SuperAdminRole,
	}
	superAdminToken, err := auth.GenerateJWT(dummySuperAdmin)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+superAdminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var users []models.AdminUser
	err = json.Unmarshal(w.Body.Bytes(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Add more tests for GetAdminUser, UpdateAdminUser, DeleteAdminUser, UpdateAdminUserStatus
// following similar patterns, ensuring to mock database interactions and check responses.

