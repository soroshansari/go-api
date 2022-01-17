package controller

import (
	errors "GoApp/src/constants/errors"
	dto "GoApp/src/dto/auth"
	"GoApp/src/model"
	"GoApp/src/provider"
	"GoApp/src/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//auth controller interface
type AuthController interface {
	Register(ctx *gin.Context)
	VerifyEmail(ctx *gin.Context)
	Login(ctx *gin.Context)
	Logout(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
}

type authController struct {
	jWtService   provider.JWTService
	configs      provider.Configs
	userService  service.UserService
	emailService provider.EmailService
	validate     validator.Validate
}

func AuthHandler(
	jWtService *provider.JWTService,
	userService *service.UserService,
	emailService *provider.EmailService,
	configs *provider.Configs,
) AuthController {
	return &authController{
		jWtService:   *jWtService,
		configs:      *configs,
		userService:  *userService,
		emailService: *emailService,
		validate:     *validator.New(),
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errors.IncorrectUserNameOrPassword})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errors.IncorrectUserNameOrPassword})
		return
	}

	if !user.Activated {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": errors.UserNotVerified})
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

// POST /api/auth/verify
// Log in the user
func (controller *authController) VerifyEmail(ctx *gin.Context) {
	var dto dto.VerifyEmail

	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	user, err := controller.userService.ActivateUser(*dto.Email, *dto.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errors.TokenExpired})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isUserExists {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": errors.UserExists})
		return
	}

	user, err := controller.userService.CreateUser(dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = controller.emailService.SendActivationEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
	})
}

// PUT /api/auth/refresh/:tokenId
// Register a user
func (controller *authController) RefreshToken(ctx *gin.Context) {
	tokenId := ctx.Param("tokenId")

	userId, err := controller.userService.FindUserIdbyRefreshToken(tokenId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"error": errors.TokenNotFound})
			return
		}
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token := controller.jWtService.GenerateToken(userId.Hex(), true, time.Minute*15)

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
	})
}

// PUT /api/auth/logout/:tokenId
// Log the user out by removing the refresh token
func (controller *authController) Logout(ctx *gin.Context) {
	tokenId := ctx.Param("tokenId")

	err := controller.userService.RemoveRefreshToken(tokenId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
	})
}
