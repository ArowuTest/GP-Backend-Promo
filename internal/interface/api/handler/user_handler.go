package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	userApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	authenticateUserService *userApp.AuthenticateUserService
	createUserService       *userApp.CreateUserService
	updateUserService       *userApp.UpdateUserService
	getUserByIDService      *userApp.GetUserService
	listUsersService        *userApp.ListUsersService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	authenticateUserService *userApp.AuthenticateUserService,
	createUserService *userApp.CreateUserService,
	updateUserService *userApp.UpdateUserService,
	getUserByIDService *userApp.GetUserService,
	listUsersService *userApp.ListUsersService,
) *UserHandler {
	return &UserHandler{
		authenticateUserService: authenticateUserService,
		createUserService:       createUserService,
		updateUserService:       updateUserService,
		getUserByIDService:      getUserByIDService,
		listUsersService:        listUsersService,
	}
}

// Login handles user authentication
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Create input with both username and email fields
	input := userApp.AuthenticateUserInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email, // Pass email if provided in the request
	}

	output, err := h.authenticateUserService.AuthenticateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Authentication failed: " + err.Error(),
		})
		return
	}

	// Return token in the nested format expected by the frontend
	// with explicit type conversions at DTO boundary
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.LoginResponse{
			Token:     output.Token,
			ExpiresAt: util.FormatTimeOrEmpty(output.ExpiresAt, time.RFC3339),
			User: response.UserResponse{
				ID:        output.ID.String(),
				Username:  output.Username,
				Email:     output.Email,
				Role:      output.Role,
				CreatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
				UpdatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
			},
		},
	})
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get creator information from context with explicit type conversion
	creatorIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Type assertion with safety check
	creatorID, ok := creatorIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid creator user ID type",
		})
		return
	}

	input := userApp.CreateUserInput{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Role:      req.Role,
		CreatedBy: creatorID,
	}

	output, err := h.createUserService.CreateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create user: " + err.Error(),
		})
		return
	}

	// Return response with explicit type conversions at DTO boundary
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  req.FullName, // Use from request since it might not be in output
			Role:      output.Role,
			CreatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
			UpdatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
		},
	})
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Parse user ID with explicit error handling
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get updater information from context with explicit type conversion
	updaterIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Type assertion with safety check
	updaterID, ok := updaterIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid updater user ID type",
		})
		return
	}

	input := userApp.UpdateUserInput{
		ID:        userID,
		Email:     req.Email,
		Role:      req.Role,
		Password:  req.Password,
		UpdatedBy: updaterID,
	}

	output, err := h.updateUserService.UpdateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update user: " + err.Error(),
		})
		return
	}

	// Return response with explicit type conversions at DTO boundary
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  req.FullName, // Use from request since it might not be in output
			Role:      output.Role,
			CreatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
			UpdatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
		},
	})
}

// GetUserByID handles retrieving a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Parse user ID with explicit error handling
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	input := userApp.GetUserInput{
		ID: userID,
	}

	output, err := h.getUserByIDService.GetUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Success: false,
			Error:   "User not found: " + err.Error(),
		})
		return
	}

	// Return response with explicit type conversions at DTO boundary
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  "", // Not available in output
			Role:      output.Role,
			CreatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
			UpdatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
		},
	})
}

// ListUsers handles retrieving a list of users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters with explicit error handling
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSizeStr := c.DefaultQuery("page_size", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	input := userApp.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := h.listUsersService.ListUsers(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list users: " + err.Error(),
		})
		return
	}

	// Convert users to response format with explicit type conversions
	users := make([]response.UserResponse, 0, len(output.Users))
	for _, u := range output.Users {
		users = append(users, response.UserResponse{
			ID:        u.ID.String(),
			Username:  u.Username,
			Email:     u.Email,
			FullName:  "", // Not available in output
			Role:      u.Role,
			CreatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
			UpdatedAt: util.FormatTimeOrEmpty(time.Now(), time.RFC3339), // Using current time as fallback
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    users,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}
