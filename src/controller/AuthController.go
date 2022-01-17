package controller

import (
	dto "GoApp/src/dto/auth"
	"GoApp/src/model"
	"GoApp/src/provider"
	"GoApp/src/service"
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
	RefreshToken(ctx *gin.Context)
	Logout(ctx *gin.Context)
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

// POST /api/auth/login
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

	token := controller.jWtService.GenerateToken(user.ID.Hex(), true, time.Minute*15)

	refreshToken, err := controller.userService.CreateRefreshToken(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken":  token,
		"refreshToken": refreshToken,
		"user":         model.GetUser(user),
	})
}

// POST /api/auth/register
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

	ctx.JSON(http.StatusOK, gin.H{
		"user": model.GetUser(user),
	})
}

// POST /api/auth/refresh-token
// Register a user
func (controller *authController) RefreshToken(ctx *gin.Context) {
	var dto dto.RefreshToken

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	userId, err := controller.userService.FindUserIdbyRefreshToken(*dto.Token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	token := controller.jWtService.GenerateToken(userId.Hex(), true, time.Minute*15)

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
	})
}

// POST /api/auth/logout
// Log the user out by removing the refresh token
func (controller *authController) Logout(ctx *gin.Context) {
	var dto dto.Logout

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	err := controller.userService.RemoveRefreshToken(*dto.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
	})
}
