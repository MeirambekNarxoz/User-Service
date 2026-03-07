package middleware

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"
)

func InternalOnlyMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Token") != expectedToken {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, gin.H{
				"error": "unauthorized internal request",
			})
			return
		}
		c.Next()
	}
}
