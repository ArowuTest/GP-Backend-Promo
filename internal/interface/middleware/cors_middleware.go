package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS for API requests
type CORSMiddleware struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           string
}

// NewCORSMiddleware creates a new CORSMiddleware
func NewCORSMiddleware(
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
	exposedHeaders []string,
	allowCredentials bool,
	maxAge string,
) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins:   allowedOrigins,
		allowedMethods:   allowedMethods,
		allowedHeaders:   allowedHeaders,
		exposedHeaders:   exposedHeaders,
		allowCredentials: allowCredentials,
		maxAge:           maxAge,
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
		
		// Set max age if provided
		if m.maxAge != "" {
			c.Writer.Header().Set("Access-Control-Max-Age", m.maxAge)
		}
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// Default returns the default CORS middleware
func Default() gin.HandlerFunc {
	middleware := NewCORSMiddleware(
		[]string{"*"},
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		[]string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		[]string{"Content-Length"},
		true,
		"43200", // 12 hours in seconds
	)
	return middleware.ApplyCORS()
}

// WithAllowedOrigins returns a CORS middleware with the specified allowed origins
func WithAllowedOrigins(origins []string) gin.HandlerFunc {
	middleware := NewCORSMiddleware(
		origins,
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		[]string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		[]string{"Content-Length"},
		true,
		"43200", // 12 hours in seconds
	)
	return middleware.ApplyCORS()
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
