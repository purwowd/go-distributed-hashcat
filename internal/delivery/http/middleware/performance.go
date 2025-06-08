package middleware

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GzipWriter wraps the ResponseWriter to provide gzip compression
type GzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

func (g *GzipWriter) WriteString(s string) (int, error) {
	return g.Writer.Write([]byte(s))
}

// Gzip middleware provides gzip compression for responses
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if client doesn't accept gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip for certain content types
		if strings.Contains(c.GetHeader("Content-Type"), "image/") ||
			strings.Contains(c.GetHeader("Content-Type"), "video/") ||
			strings.Contains(c.GetHeader("Content-Type"), "application/octet-stream") {
			c.Next()
			return
		}

		// Create gzip writer
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		// Set headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Wrap writer
		c.Writer = &GzipWriter{
			ResponseWriter: c.Writer,
			Writer:         gz,
		}

		c.Next()
	}
}

// Cache middleware adds appropriate cache headers for static content
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Set cache headers based on content type
		if strings.HasPrefix(path, "/static/") {
			// Cache static assets for 1 hour
			c.Header("Cache-Control", "public, max-age=3600")
			c.Header("ETag", `"`+time.Now().Format("20060102150405")+`"`)
		} else if strings.HasPrefix(path, "/api/v1/") {
			// API responses - short cache for GET requests
			if c.Request.Method == "GET" {
				c.Header("Cache-Control", "public, max-age=30")
			} else {
				c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			}
		} else {
			// Default - no cache for dynamic content
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		c.Next()
	}
}

// SecurityHeaders adds security-related headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only add HSTS for HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RequestTimeout middleware sets a timeout for requests
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a deadline for the request context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Check if the context was canceled due to timeout
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
			})
			c.Abort()
		}
	}
}

// Performance middleware combines multiple performance optimizations
func Performance() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Set performance-related headers
		c.Header("Server", "HashcatServer/1.0")

		// Measure request processing time
		start := time.Now()

		c.Next()

		// Add processing time header for debugging
		processingTime := time.Since(start)
		c.Header("X-Processing-Time", processingTime.String())
	})
}
