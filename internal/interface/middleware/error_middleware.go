package middleware

import (
	"github.com/gin-gonic/gin"
)

// ErrorMiddleware handles error responses for API requests
type ErrorMiddleware struct {
	debug bool
}

// NewErrorMiddleware creates a new ErrorMiddleware
func NewErrorMiddleware(debug bool) *ErrorMiddleware {
	return &ErrorMiddleware{
		debug: debug,
	}
}

// HandleErrors handles errors in the request pipeline
func (m *ErrorMiddleware) HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Determine status code
			statusCode := c.Writer.Status()
			if statusCode == 200 {
				statusCode = 500
			}

			// Prepare error response
			errorResponse := gin.H{
				"error": err.Error(),
			}

			// Add debug information if enabled
			if m.debug {
				errorResponse["debug"] = err.Meta
			}

			// Send error response
			c.JSON(statusCode, errorResponse)
		}
	}
}
