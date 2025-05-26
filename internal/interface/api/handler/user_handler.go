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
	createUserService *userApp.CreateUserService
	updateUserService *userApp.UpdateUserService
	getUserService *userApp.GetUserService
	listUsersService *userApp.ListUsersService
	authenticateUserService *userApp.AuthenticateUserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	createUserService *userApp.CreateUserService,
	updateUserService *userApp.UpdateUserService,
	getUserService *userApp.GetUserService,
	listUsersService *userApp.ListUsersService,
	authenticateUserService *userApp.AuthenticateUserService,
) *UserHandler {
	return &UserHandler{
		createUserService: createUserService,
		updateUserService: updateUserService,
		getUserService: getUserService,
		listUsersService: listUsersService,
		authenticateUserService: authenticateUserService,
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

	input := userApp.AuthenticateUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.authenticateUserService.AuthenticateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Authentication failed: " + err.Error(),
		})
		return
	}

	// Return response with explicit type conversions at DTO boundary
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.LoginResponse{
			Token: output.Token,
			ExpiresAt: util.FormatTimeOrEmpty(output.ExpiresAt, time.RFC3339),
			User: response.UserResponse{
				ID:       output.User.ID.String(),
				Username: output.User.Username,
				Email:    output.User.Email,
				Role:     output.User.Role,
				IsActive: true, // Default to true since field is missing
				// Include additional fields for frontend compatibility
				FullName:  output.User.Username, // Use username as fallback for fullname
				CreatedAt: time.Now().Format(time.RFC3339), // Default since field is missing
				UpdatedAt: time.Now().Format(time.RFC3339), // Default since field is missing
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
	var creatorID uuid.UUID
	switch id := creatorIDValue.(type) {
	case uuid.UUID:
		creatorID = id
	case string:
		var err error
		creatorID, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid creator user ID format",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid creator user ID type",
		})
		return
	}

	input := userApp.CreateUserInput{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
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
			ID:       output.ID.String(),
			Username: output.Username,
			Email:    output.Email,
			Role:     output.Role,
			IsActive: output.IsActive,
			// Include additional fields for frontend compatibility
			FullName:  req.Username, // Use username as fallback for fullname
			CreatedAt: util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
			UpdatedAt: util.FormatTimeOrEmpty(output.UpdatedAt, time.RFC3339),
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
	var updaterID uuid.UUID
	switch id := updaterIDValue.(type) {
	case uuid.UUID:
		updaterID = id
	case string:
		var err error
		updaterID, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid updater user ID format",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid updater user ID type",
		})
		return
	}
	
	input := userApp.UpdateUserInput{
		ID:        userID,
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		IsActive:  req.IsActive,
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
			ID:       output.ID.String(),
			Username: output.Username,
			Email:    output.Email,
			Role:     output.Role,
			IsActive: output.IsActive,
			// Include additional fields for frontend compatibility
			FullName:  req.Username, // Use username as fallback for fullname
			CreatedAt: util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
			UpdatedAt: util.FormatTimeOrEmpty(output.UpdatedAt, time.RFC3339),
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
	
	output, err := h.getUserService.GetUser(c.Request.Context(), input)
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
			ID:       output.ID.String(),
			Username: output.Username,
			Email:    output.Email,
			Role:     output.Role,
			IsActive: output.IsActive,
			// Include additional fields for frontend compatibility
			FullName:  output.Username, // Use username as fallback for fullname
			CreatedAt: util.FormatTimeOrEmpty(output.CreatedAt, time.RFC3339),
			UpdatedAt: util.FormatTimeOrEmpty(output.UpdatedAt, time.RFC3339),
		},
	})
}

// ListUsers handles retrieving a list of users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters with explicit error handling
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
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
			ID:       u.ID.String(),
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role,
			IsActive: u.IsActive,
			// Include additional fields for frontend compatibility
			FullName:  u.Username, // Use username as fallback for fullname
			CreatedAt: util.FormatTimeOrEmpty(u.CreatedAt, time.RFC3339),
			UpdatedAt: util.FormatTimeOrEmpty(u.UpdatedAt, time.RFC3339),
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
