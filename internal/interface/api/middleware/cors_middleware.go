package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposeHeaders    []string
	allowCredentials bool
}

// NewCORSMiddleware creates a new CORSMiddleware
func NewCORSMiddleware(
	allowOrigins []string,
	allowMethods []string,
	allowHeaders []string,
	exposeHeaders []string,
	allowCredentials bool,
) *CORSMiddleware {
	return &CORSMiddleware{
		allowOrigins:     allowOrigins,
		allowMethods:     allowMethods,
		allowHeaders:     allowHeaders,
		exposeHeaders:    exposeHeaders,
		allowCredentials: allowCredentials,
	}
}

// Default creates a new CORSMiddleware with default settings
func Default() *CORSMiddleware {
	return &CORSMiddleware{
		allowOrigins:     []string{"*"},
		allowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		allowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		exposeHeaders:    []string{"Content-Length"},
		allowCredentials: false,
	}
}

// Handle returns a gin handler function for CORS
func (m *CORSMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		origin := c.Request.Header.Get("Origin")
		
		// Check if the origin is allowed
		allowOrigin := "*"
		if len(m.allowOrigins) > 0 && m.allowOrigins[0] != "*" {
			allowOrigin = ""
			for _, o := range m.allowOrigins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
		}
		
		// Set Access-Control-Allow-Origin
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		
		// Set Access-Control-Allow-Methods
		if len(m.allowMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(m.allowMethods))
		}
		
		// Set Access-Control-Allow-Headers
		if len(m.allowHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(m.allowHeaders))
		}
		
		// Set Access-Control-Expose-Headers
		if len(m.exposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(m.exposeHeaders))
		}
		
		// Set Access-Control-Allow-Credentials
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
	if len(strings) == 0 {
		return ""
	}
	
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += ", " + strings[i]
	}
	
	return result
}
