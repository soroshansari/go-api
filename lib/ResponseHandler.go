package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const IncorrectUserNameOrPassword = "IncorrectUserNameOrPassword"
const UserNotVerified = "UserNotVerified"
const UserNotFound = "UserNotFound"
const UserAlreadyActivated = "UserAlreadyActivated"
const UserExists = "UserExists"
const TokenExpired = "TokenExpired"
const TokenNotFound = "TokenNotFound"
const IncorrectOldPassword = "IncorrectOldPassword"

func JsonResponse(c *gin.Context, data interface{}) {
	if data != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "Success",
			"data":   data,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": "Success",
		})
	}
}

func ErrorResponse(c *gin.Context, httpStatus int, err string) {
	if err == "" {
		c.AbortWithStatusJSON(httpStatus, gin.H{
			"status": "Failed",
		})
	} else {
		c.AbortWithStatusJSON(httpStatus, gin.H{
			"status": "Failed",
			"error":  err,
		})
	}
}
