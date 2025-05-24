package request

// AdminResetPasswordRequest defines the request for an admin resetting a user's password
type AdminResetPasswordRequest struct {
	UserID      string `json:"userId" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
