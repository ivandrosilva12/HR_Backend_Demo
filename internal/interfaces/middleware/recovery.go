package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Prevents crashes due to panics and returns 500 gracefully.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Recovered from panic:", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
		}()
		c.Next()
	}
}
