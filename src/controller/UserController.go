package controller

import (
	errors "GoApp/src/constants/errors"
	dto "GoApp/src/dto/user"
	"GoApp/src/lib"
	"GoApp/src/model"
	"GoApp/src/provider"
	"GoApp/src/service"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//auth controller interface
type UserController interface {
	Me(ctx *gin.Context)
	ChangePassword(ctx *gin.Context)
	UploadProfile(ctx *gin.Context)
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
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnauthorized, "")
		return
	}

	lib.JsonResponse(ctx, model.GetUser(user, &controller.configs))
}

// POST /api/user/change-password
func (controller *userController) ChangePassword(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)
	var dto dto.ChangePassword

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindById(userId)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnauthorized, "")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.OldPassword)); err != nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.IncorrectOldPassword)
		return
	}

	err = controller.userService.UpdatePassword(user.ID, *dto.NewPassword)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, nil)
}

func (controller *userController) UploadProfile(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)

	user, err := controller.userService.FindById(userId)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnauthorized, "")
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}
	fileExtension := filepath.Ext(header.Filename)
	filename := uuid.NewString() + fileExtension
	out, err := os.Create("public/profile/" + filename)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if user.Profile != "" {
		err = os.Remove("public/profile/" + user.Profile)
		if err != nil {
			log.Default().Println(err.Error())
		}
	}

	err = controller.userService.UpdateProfile(user.ID, filename)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	filepath := controller.configs.Domain + "/public/profile/" + filename
	lib.JsonResponse(ctx, gin.H{"filepath": filepath})
}
