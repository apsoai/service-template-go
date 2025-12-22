package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides basic authentication middleware
// Customize this for your authentication needs
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add your authentication logic here
		// Example: Check for API key, JWT token, etc.

		// For now, just pass through
		c.Next()
	}
}

// RequireAuth is a middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add your authentication check here
		// If not authenticated, return 401

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Add token validation logic here
		c.Next()
	}
}
