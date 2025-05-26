package response

// DeleteConfirmationResponse represents a response confirming deletion
type DeleteConfirmationResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}
