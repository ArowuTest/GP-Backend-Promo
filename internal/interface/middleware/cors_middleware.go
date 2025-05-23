package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS for API requests
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
	exposedHeaders []string
	allowCredentials bool
}

// NewCORSMiddleware creates a new CORSMiddleware
func NewCORSMiddleware(
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
	exposedHeaders []string,
	allowCredentials bool,
) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
		allowedMethods: allowedMethods,
		allowedHeaders: allowedHeaders,
		exposedHeaders: exposedHeaders,
		allowCredentials: allowCredentials,
	}
}

// ApplyCORS applies CORS headers to responses
func (m *CORSMiddleware) ApplyCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if the origin is allowed
		allowed := false
		for _, allowedOrigin := range m.allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(m.allowedOrigins) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Origin", m.allowedOrigins[0])
		}
		
		// Set allowed methods
		if len(m.allowedMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(m.allowedMethods))
		}
		
		// Set allowed headers
		if len(m.allowedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(m.allowedHeaders))
		}
		
		// Set exposed headers
		if len(m.exposedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(m.exposedHeaders))
		}
		
		// Set allow credentials
		if m.allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// Helper function to join strings with comma
func joinStrings(strings []string) string {
	result := ""
	for i, s := range strings {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
