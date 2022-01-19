package middlewares

import (
	"GoApp/lib"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqKey := c.Request.Header.Get("X-Auth-Key")

		if key != reqKey {
			lib.ErrorResponse(c, http.StatusUnauthorized, "Invalid auth key or secret")
			return
		}
		c.Next()
	}
}
