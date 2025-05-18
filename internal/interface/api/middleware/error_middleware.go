package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// ErrorMiddleware handles error responses
type ErrorMiddleware struct {
	isDevelopment bool
}

// NewErrorMiddleware creates a new ErrorMiddleware
func NewErrorMiddleware(isDevelopment bool) *ErrorMiddleware {
	return &ErrorMiddleware{
		isDevelopment: isDevelopment,
	}
}

// Handle returns a gin handler function for error handling
func (m *ErrorMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			
			// Prepare error response
			statusCode := http.StatusInternalServerError
			errorMessage := "Internal Server Error"
			errorDetails := err.Error()
			
			// Check for specific error types
			switch err.Type {
			case gin.ErrorTypeBind:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid Request"
			case gin.ErrorTypePublic:
				statusCode = http.StatusBadRequest
				errorMessage = err.Error()
				errorDetails = ""
			case gin.ErrorTypePrivate:
				// Keep default status code and message
			case gin.ErrorTypeRender:
				statusCode = http.StatusInternalServerError
				errorMessage = "Render Error"
			}
			
			// Create error response
			errResponse := response.ErrorResponse{
				Success: false,
				Error:   errorMessage,
			}
			
			// Add details in development mode
			if m.isDevelopment {
				errResponse.Details = errorDetails
			}
			
			// Send error response
			c.JSON(statusCode, errResponse)
		}
	}
}

// Recovery returns a gin handler function for panic recovery
func (m *ErrorMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the stack trace
				stackTrace := debug.Stack()
				
				// Prepare error response
				errResponse := response.ErrorResponse{
					Success: false,
					Error:   "Internal Server Error",
				}
				
				// Add details in development mode
				if m.isDevelopment {
					errResponse.Details = string(stackTrace)
				}
				
				// Send error response
				c.JSON(http.StatusInternalServerError, errResponse)
				
				// Abort the request
				c.Abort()
			}
		}()
		
		c.Next()
	}
}
