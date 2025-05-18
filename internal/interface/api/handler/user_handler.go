package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	authenticateUserUseCase *user.AuthenticateUserUseCase
	createUserUseCase       *user.CreateUserUseCase
	updateUserUseCase       *user.UpdateUserUseCase
	getUserByIDUseCase      *user.GetUserUseCase
	listUsersUseCase        *user.ListUsersUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	authenticateUserUseCase *user.AuthenticateUserUseCase,
	createUserUseCase *user.CreateUserUseCase,
	updateUserUseCase *user.UpdateUserUseCase,
	getUserByIDUseCase *user.GetUserUseCase,
	listUsersUseCase *user.ListUsersUseCase,
) *UserHandler {
	return &UserHandler{
		authenticateUserUseCase: authenticateUserUseCase,
		createUserUseCase:       createUserUseCase,
		updateUserUseCase:       updateUserUseCase,
		getUserByIDUseCase:      getUserByIDUseCase,
		listUsersUseCase:        listUsersUseCase,
	}
}

// Login handles user authentication
func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	input := user.AuthenticateUserInput{
		Username: request.Username,
		Password: request.Password,
	}

	output, err := h.authenticateUserUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Authentication failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"token":      output.Token,
			"expires_at": output.ExpiresAt,
			"user": gin.H{
				"id":        output.User.ID,
				"username":  output.User.Username,
				"email":     output.User.Email,
				"full_name": output.User.FullName,
				"role":      output.User.Role,
			},
		},
	})
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		FullName string `json:"full_name" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Get creator information from context
	creatorID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User information not found",
		})
		return
	}

	input := user.CreateUserInput{
		Username:  request.Username,
		Password:  request.Password,
		Email:     request.Email,
		FullName:  request.FullName,
		Role:      request.Role,
		CreatedBy: creatorID.(string),
	}

	output, err := h.createUserUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":        output.User.ID,
			"username":  output.User.Username,
			"email":     output.User.Email,
			"full_name": output.User.FullName,
			"role":      output.User.Role,
			"active":    output.User.Active,
		},
	})
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Details: "User ID is required",
		})
		return
	}

	var request struct {
		Email    string `json:"email" binding:"omitempty,email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		Active   *bool  `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Get updater information from context
	updaterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User information not found",
		})
		return
	}

	input := user.UpdateUserInput{
		UserID:    userID,
		Email:     request.Email,
		FullName:  request.FullName,
		Role:      request.Role,
		Active:    request.Active,
		UpdatedBy: updaterID.(string),
	}

	output, err := h.updateUserUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":        output.User.ID,
			"username":  output.User.Username,
			"email":     output.User.Email,
			"full_name": output.User.FullName,
			"role":      output.User.Role,
			"active":    output.User.Active,
		},
	})
}

// GetUserByID handles retrieving a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Details: "User ID is required",
		})
		return
	}

	input := user.GetUserInput{
		UserID: userID,
	}

	output, err := h.getUserByIDUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Success: false,
			Error:   "User not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":        output.User.ID,
			"username":  output.User.Username,
			"email":     output.User.Email,
			"full_name": output.User.FullName,
			"role":      output.User.Role,
			"active":    output.User.Active,
		},
	})
}

// ListUsers handles retrieving a list of users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	input := user.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := h.listUsersUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list users",
			Details: err.Error(),
		})
		return
	}

	// Convert users to response format
	users := make([]gin.H, len(output.Users))
	for i, u := range output.Users {
		users[i] = gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"email":     u.Email,
			"full_name": u.FullName,
			"role":      u.Role,
			"active":    u.Active,
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"users":      users,
			"total":      output.Total,
			"page":       output.Page,
			"page_size":  output.PageSize,
			"total_pages": (output.Total + int64(output.PageSize) - 1) / int64(output.PageSize),
		},
	})
}
