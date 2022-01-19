package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JsonResponse(ctx *gin.Context, data interface{}) {
	if data != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "Success",
			"data":   data,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "Success",
		})
	}
}

func ErrorResponse(ctx *gin.Context, httpStatus int, err string) {
	if err == "" {
		ctx.JSON(httpStatus, gin.H{
			"status": "Failed",
		})
	} else {
		ctx.JSON(httpStatus, gin.H{
			"status": "Failed",
			"error":  err,
		})
	}
}
