package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
)

// ResetPasswordHandler handles password reset requests
type ResetPasswordHandler struct {
	resetPasswordService *user.ResetPasswordService
}

// NewResetPasswordHandler creates a new ResetPasswordHandler
func NewResetPasswordHandler(resetPasswordService *user.ResetPasswordService) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		resetPasswordService: resetPasswordService,
	}
}

// ResetPasswordRequest defines the request body for resetting a password
type ResetPasswordRequest struct {
	UserID      string `json:"user_id" binding:"required,uuid"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Handle handles the reset password request
func (h *ResetPasswordHandler) Handle(c *gin.Context) {
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
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user not authenticated"})
		return
	}
	
	// Reset password
	input := user.ResetPasswordInput{
		UserID:      userID,
		NewPassword: req.NewPassword,
		AdminUserID: adminUserID.(uuid.UUID),
	}
	
	output, err := h.resetPasswordService.ResetPassword(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password: " + err.Error()})
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Password reset successfully",
		"user_id":  output.UserID.String(),
		"username": output.Username,
		"email":    output.Email,
	})
}
