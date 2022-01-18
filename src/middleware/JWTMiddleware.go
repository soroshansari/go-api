package middleware

import (
	"GoApp/src/provider"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthorizeJWT(configs *provider.Configs) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if isBearer := strings.HasPrefix(authHeader, BEARER_SCHEMA); !isBearer {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := authHeader[len(BEARER_SCHEMA):]
		if tokenString == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token, err := provider.JWTAuthService(configs).ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		if !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		if !(len(claims["sub"].(string)) > 0) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userId", claims["sub"])
	}
}
