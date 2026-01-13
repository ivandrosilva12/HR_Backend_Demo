package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func StripTrailingSlash() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.URL.Path = strings.TrimSuffix(c.Request.URL.Path, "/")
		c.Next()
	}
}
