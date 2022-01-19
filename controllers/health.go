package controllers

import (
	"GoApp/lib"

	"github.com/gin-gonic/gin"
)

//auth controllers interface
type HealthController interface {
	Status(c *gin.Context)
}

type healthController struct {
}

func HealthControllerHandler() HealthController {
	return &healthController{}
}

// GET /
func (controller *healthController) Status(c *gin.Context) {
	lib.JsonResponse(c, nil)
}
