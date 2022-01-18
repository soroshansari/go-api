package controller

import (
	errors "GoApp/src/constants/errors"
	dto "GoApp/src/dto/user"
	"GoApp/src/model"
	"GoApp/src/provider"
	"GoApp/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

//auth controller interface
type UserController interface {
	Me(ctx *gin.Context)
	ChangePassword(ctx *gin.Context)
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

// GET /api/user/me
// get authenticated user info
func (controller *userController) Me(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)

	user, err := controller.userService.FindById(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, model.GetUser(user))
}

// POST /api/user/change-password
func (controller *userController) ChangePassword(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)
	var dto dto.ChangePassword

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	user, err := controller.userService.FindById(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.OldPassword)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errors.IncorrectOldPassword})
		return
	}

	err = controller.userService.UpdatePassword(userId, *dto.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
	})
}
