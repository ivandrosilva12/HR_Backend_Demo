package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BodySizeLimit(limitBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Request.Body = http.MaxBytesReader(writer, body, size)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limitBytes)

		if err := c.Request.ParseForm(); err != nil && err.Error() == "http: request body too large" {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Payload muito grande. O limite é " + strconv.FormatInt(limitBytes, 10) + " bytes",
			})
			return
		}

		c.Next()
	}
}
