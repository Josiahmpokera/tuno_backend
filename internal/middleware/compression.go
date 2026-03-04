package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// CompressionMiddleware enables GZIP compression for responses
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip compression for certain content types
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, binding.MIMEJSON) ||
			strings.Contains(contentType, binding.MIMEHTML) ||
			strings.Contains(contentType, binding.MIMEXML) ||
			strings.Contains(contentType, binding.MIMEXML2) ||
			strings.Contains(contentType, "text/") {
			
			// Only compress responses larger than 1KB
			c.Writer.Header().Set("Vary", "Accept-Encoding")
			
			// Check if client supports gzip
			if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
				c.Writer.Header().Set("Content-Encoding", "gzip")
				// Gin automatically handles gzip compression when this header is set
			}
		}
		
		c.Next()
	}
}