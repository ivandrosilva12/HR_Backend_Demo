package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateInput(schema interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(&schema); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		c.Next()
	}
}
