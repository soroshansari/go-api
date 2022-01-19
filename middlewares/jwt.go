package middlewares

import (
	"GoApp/lib"
	"GoApp/providers"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthorizeJWT(jwtService providers.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			lib.ErrorResponse(c, http.StatusUnauthorized, "")
			return
		}
		if isBearer := strings.HasPrefix(authHeader, BEARER_SCHEMA); !isBearer {

			lib.ErrorResponse(c, http.StatusUnauthorized, "")
			return
		}
		tokenString := authHeader[len(BEARER_SCHEMA):]
		if tokenString == "" {

			lib.ErrorResponse(c, http.StatusUnauthorized, "")
			return
		}
		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			lib.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		if !token.Valid {

			lib.ErrorResponse(c, http.StatusUnauthorized, "")
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		if !(len(claims["sub"].(string)) > 0) {

			lib.ErrorResponse(c, http.StatusUnauthorized, "")
			return
		}
		c.Set("userId", claims["sub"])
	}
}
