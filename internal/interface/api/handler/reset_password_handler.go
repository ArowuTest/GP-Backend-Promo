package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/api/middleware"
)

// ResetPasswordRequest defines the request body for resetting a user's password
type ResetPasswordRequest struct {
	UserID      string `json:"user_id" binding:"required,uuid"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordResponse defines the response body for the reset password endpoint
type ResetPasswordResponse struct {
	Message  string `json:"message"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ResetPasswordHandler handles the reset password endpoint
type ResetPasswordHandler struct {
	userService *user.UserService
}

// NewResetPasswordHandler creates a new ResetPasswordHandler
func NewResetPasswordHandler(userService *user.UserService) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		userService: userService,
	}
}

// Handle handles the reset password request
func (h *ResetPasswordHandler) Handle(c *gin.Context) {
	// Check if user has admin privileges
	isAdmin, err := h.userService.IsAdminUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user privileges"})
		return
	}
	
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient privileges. Only SUPER_ADMIN and ADMIN roles can reset passwords."})
		return
	}
	
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	
	// Get admin user ID from context
	adminUserID := h.userService.GetCurrentUserID(c.Request.Context())
	if adminUserID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user not authenticated"})
		return
	}
	
	// Reset password
	input := user.ResetPasswordInput{
		UserID:      userID,
		NewPassword: req.NewPassword,
		AdminUserID: adminUserID,
	}
	
	output, err := h.userService.ResetPassword(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password: " + err.Error()})
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, ResetPasswordResponse{
		Message:  "Password reset successfully",
		UserID:   output.UserID.String(),
		Username: output.Username,
		Email:    output.Email,
	})
}
