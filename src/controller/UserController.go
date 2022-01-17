package controller

import (
	"GoApp/src/model"
	"GoApp/src/provider"
	"GoApp/src/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

//auth controller interface
type UserController interface {
	Me(ctx *gin.Context)
}

type userController struct {
	configs     provider.Configs
	userService service.UserService
	validate    validator.Validate
}

func UserHandler(
	userService *service.UserService,
	configs *provider.Configs,
) UserController {
	return &userController{
		configs:     *configs,
		userService: *userService,
		validate:    *validator.New(),
	}
}

func (controller *userController) Me(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)
	fmt.Print("userId:")
	fmt.Println(userId)

	user, err := controller.userService.FindById(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if user == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": model.GetUser(user)})
}
