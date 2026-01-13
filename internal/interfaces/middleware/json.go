// middleware/json_validation.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireJSON rejects non-JSON requests early.
func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() != "application/json" {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json",
			})
			return
		}
		c.Next()
	}
}

// ValidateJSON binds & validates a JSON body to T and stores it in context.
// Use struct tags like `binding:"required,email"` on T's fields.
func ValidateJSON[T any](ctxKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload T
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"details": err.Error(),
			})
			return
		}
		c.Set(ctxKey, payload)
		c.Next()
	}
}
