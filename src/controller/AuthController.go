package controller

import (
	dto "RentBuddi/src/dto/auth"
	"RentBuddi/src/model"
	"RentBuddi/src/provider"
	"RentBuddi/src/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

//auth controller interface
type AuthController interface {
	Login(ctx *gin.Context)
	Register(ctx *gin.Context)
	// Logout(ctx *gin.Context)
}

type authController struct {
	jWtService  provider.JWTService
	configs     provider.Configs
	userService service.UserService
	validate    validator.Validate
}

func AuthHandler(
	jWtService *provider.JWTService,
	userService *service.UserService,
	configs *provider.Configs,
) AuthController {
	return &authController{
		jWtService:  *jWtService,
		configs:     *configs,
		userService: *userService,
		validate:    *validator.New(),
	}
}

// GET /api/auth/login
// Log in the user
func (controller *authController) Login(ctx *gin.Context) {
	var dto dto.LoginCredentials

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	user, err := controller.userService.FindUser(*dto.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect email or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect email or password"})
		return
	}

	token := controller.jWtService.GenerateToken(user.User_id, true, time.Hour*24*30)
	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        model.GetUser(user),
	})
}

// GET /api/auth/register
// Register a user
func (controller *authController) Register(ctx *gin.Context) {
	var dto dto.RegisterCredentials

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	isUserExists, err := controller.userService.UserExists(*dto.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if isUserExists {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "User already exists"})
		return
	}

	user, err := controller.userService.CreateUser(dto)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	token := controller.jWtService.GenerateToken(user.User_id, true, time.Hour*24*30)
	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        model.GetUser(user),
	})
}
