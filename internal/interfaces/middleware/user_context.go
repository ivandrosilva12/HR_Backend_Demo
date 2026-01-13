package middleware

import "github.com/gin-gonic/gin"

func UserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse token, fetch user from DB, attach to context
		var user string
		c.Set("user", user)
		c.Next()
	}
}
