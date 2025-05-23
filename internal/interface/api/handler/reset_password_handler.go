package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
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
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	// Parse user ID with explicit error handling
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}
	
	// Get admin user ID from context with explicit type conversion
	adminUserIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Admin user not authenticated",
		})
		return
	}
	
	// Type assertion with safety check
	adminUserID, ok := adminUserIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid admin user ID type",
		})
		return
	}
	
	// Reset password with explicit input structure
	input := user.ResetPasswordInput{
		UserID:      userID,
		NewPassword: req.NewPassword,
		AdminUserID: adminUserID,
	}
	
	output, err := h.resetPasswordService.ResetPassword(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to reset password: " + err.Error(),
		})
		return
	}
	
	// Return success response using standard response format
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: map[string]string{
			"message":  "Password reset successfully",
			"user_id":  output.UserID.String(),
			"username": output.Username,
			"email":    output.Email,
		},
	})
}
