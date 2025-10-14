package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Debug: Log the request
		fmt.Printf("DEBUG: CORS middleware called for %s %s\n", c.Request.Method, c.Request.URL.Path)
		
		// Force wildcard CORS for development
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Origin, Content-Type, Accept, Authorization, X-Requested-With, Cache-Control, Content-Length, Accept-Encoding, X-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Debug: Log the headers being set
		fmt.Printf("DEBUG: Setting Access-Control-Allow-Origin to: *\n")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Origin, Content-Type, Accept, Authorization, X-Requested-With, Cache-Control, Content-Length, Accept-Encoding, X-Token")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT, OPTIONS")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CORSWithSpecificOrigin creates CORS middleware for specific origin
func CORSWithSpecificOrigin(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Debug: Log the allowed origin
		fmt.Printf("DEBUG: CORS middleware using origin: %s\n", allowedOrigin)
		
		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Origin, Content-Type, Accept, Authorization, X-Requested-With, Cache-Control, Content-Length, Accept-Encoding, X-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Origin, Content-Type, Accept, Authorization, X-Requested-With, Cache-Control, Content-Length, Accept-Encoding, X-Token")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT, OPTIONS")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
