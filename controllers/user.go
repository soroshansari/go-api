package controllers

import (
	"GoApp/db"
	dto "GoApp/dto/user"
	"GoApp/lib"
	"GoApp/models"
	"GoApp/providers"
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

//auth controllers interface
type UserController interface {
	Me(c *gin.Context)
	ChangePassword(c *gin.Context)
	UploadProfile(c *gin.Context)
	UpdateUserDetails(c *gin.Context)
}

type userController struct {
	configs     providers.Config
	userService db.UserService
	validate    validator.Validate
}

func UserHandler(
	userService *db.UserService,
	configs *providers.Config,
) UserController {
	return &userController{
		configs:     *configs,
		userService: *userService,
		validate:    *validator.New(),
	}
}

// GET /api/user/details
// get authenticated user info
func (controller *userController) Me(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	user, err := controller.userService.FindById(userId)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnauthorized, "")
		return
	}

	lib.JsonResponse(c, models.GetUser(user, &controller.configs))
}

// POST /api/user/details
// get authenticated user info
func (controller *userController) UpdateUserDetails(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	var dto dto.UpdateUserDetails

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	err := controller.userService.UpdateDetail(userId, *dto.Firstname, *dto.Lastname)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

// POST /api/user/change-password
func (controller *userController) ChangePassword(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	var dto dto.ChangePassword

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindById(userId)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnauthorized, "")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.OldPassword)); err != nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.IncorrectOldPassword)
		return
	}

	err = controller.userService.UpdatePassword(user.ID, *dto.NewPassword)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

func (controller *userController) UploadProfile(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	user, err := controller.userService.FindById(userId)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnauthorized, "")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	fileExtension := filepath.Ext(header.Filename)
	filename := uuid.NewString() + fileExtension
	out, err := os.Create("public/profile/" + filename)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
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
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	filepath := controller.configs.Domain + "/public/profile/" + filename
	lib.JsonResponse(c, gin.H{"filepath": filepath})
}
