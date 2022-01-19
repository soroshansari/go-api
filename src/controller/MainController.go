package controller

import (
	"GoApp/src/lib"

	"github.com/gin-gonic/gin"
)

//auth controller interface
type MainController interface {
	HealthCheck(c *gin.Context)
}

type mainController struct {
}

func MainControllerHandler() MainController {
	return &mainController{}
}

// GET /
func (controller *mainController) HealthCheck(ctx *gin.Context) {
	lib.JsonResponse(ctx, nil)
}
