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

	// FIXED: Return token at the top level instead of nested in a data object
	// This matches the format expected by the frontend
	c.JSON(http.StatusOK, gin.H{
		"token":    output.Token,
		"user_id":  output.ID.String(),
		"username": output.Username,
		"role":     output.Role,
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

	// Get creator information from context
	creatorID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	input := userApp.CreateUserInput{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Role:      req.Role,
		CreatedBy: creatorID.(uuid.UUID),
	}

	output, err := h.createUserService.CreateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  req.FullName, // Use from request since it might not be in output
			Role:      output.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(c *gin.Context) {
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

	// Get updater information from context
	updaterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	input := userApp.UpdateUserInput{
		ID:        userID,
		Email:     req.Email,
		Role:      req.Role,
		Password:  req.Password,
		UpdatedBy: updaterID.(uuid.UUID),
	}

	output, err := h.updateUserService.UpdateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  req.FullName, // Use from request since it might not be in output
			Role:      output.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// GetUserByID handles retrieving a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
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

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.UserResponse{
			ID:        output.ID.String(),
			Username:  output.Username,
			Email:     output.Email,
			FullName:  "", // Not available in output
			Role:      output.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// ListUsers handles retrieving a list of users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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

	// Convert users to response format
	users := make([]response.UserResponse, 0, len(output.Users))
	for _, u := range output.Users {
		users = append(users, response.UserResponse{
			ID:        u.ID.String(),
			Username:  u.Username,
			Email:     u.Email,
			FullName:  "", // Not available in output
			Role:      u.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
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
