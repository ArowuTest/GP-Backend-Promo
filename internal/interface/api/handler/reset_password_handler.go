package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	userApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ResetPasswordHandler handles password reset HTTP requests
type ResetPasswordHandler struct {
	resetPasswordService *userApp.ResetPasswordService
}

// NewResetPasswordHandler creates a new ResetPasswordHandler
func NewResetPasswordHandler(
	resetPasswordService *userApp.ResetPasswordService,
) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		resetPasswordService: resetPasswordService,
	}
}

// ResetPassword handles POST /api/admin/users/reset-password
func (h *ResetPasswordHandler) ResetPassword(c *gin.Context) {
	var req request.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context with explicit type conversion
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	var userID uuid.UUID
	switch id := userIDValue.(type) {
	case uuid.UUID:
		userID = id
	case string:
		var err error
		userID, err = uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID format in token",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID type in token",
		})
		return
	}

	// Parse target user ID
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	input := userApp.ResetPasswordInput{
		UserID:      targetUserID,
		NewPassword: req.NewPassword,
		AdminUserID: userID,
	}

	_, err = h.resetPasswordService.ResetPassword(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Failed to reset password: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}
