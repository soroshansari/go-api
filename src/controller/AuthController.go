package controller

import (
	errors "GoApp/src/constants/errors"
	dto "GoApp/src/dto/auth"
	"GoApp/src/lib"
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
	ForgotPass(ctx *gin.Context)
	ResendActivationEmail(ctx *gin.Context)
	ResetPass(ctx *gin.Context)
}

type authController struct {
	jWtService          provider.JWTService
	configs             provider.Configs
	userService         service.UserService
	refreshTokenService service.RefreshTokenService
	emailService        provider.EmailService
	validate            validator.Validate
}

func AuthHandler(
	jWtService *provider.JWTService,
	userService *service.UserService,
	refreshTokenService *service.RefreshTokenService,
	emailService *provider.EmailService,
	configs *provider.Configs,
) AuthController {
	return &authController{
		jWtService:          *jWtService,
		configs:             *configs,
		userService:         *userService,
		refreshTokenService: *refreshTokenService,
		emailService:        *emailService,
		validate:            *validator.New(),
	}
}

// POST /api/auth/login
// Log in the user
func (controller *authController) Login(ctx *gin.Context) {
	var dto dto.LoginCredentials

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindUser(*dto.Email)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.IncorrectUserNameOrPassword)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.Password)); err != nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.IncorrectUserNameOrPassword)
		return
	}

	if !user.Activated {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.UserNotVerified)
		return
	}

	token := controller.jWtService.GenerateToken(user.ID.Hex(), true, time.Minute*15)

	refreshToken, err := controller.refreshTokenService.CreateRefreshToken(user.ID)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, gin.H{
		"accessToken":  token,
		"refreshToken": refreshToken,
		"user":         model.GetUser(user, &controller.configs),
	})
}

// POST /api/auth/verify
func (controller *authController) VerifyEmail(ctx *gin.Context) {
	var dto dto.VerifyEmail

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.ActivateUser(*dto.Email, *dto.Code, "")
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.TokenExpired)
		return
	}

	lib.JsonResponse(ctx, nil)
}

// POST /api/auth/register
// Register a user
func (controller *authController) Register(ctx *gin.Context) {
	var dto dto.RegisterCredentials

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	isUserExists, err := controller.userService.UserExists(*dto.Email)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if isUserExists {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.UserExists)
		return
	}

	user, err := controller.userService.CreateUser(dto)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	err = controller.emailService.SendActivationEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, nil)
}

// PUT /api/auth/refresh/:tokenId
// Register a user
func (controller *authController) RefreshToken(ctx *gin.Context) {
	tokenId := ctx.Param("tokenId")

	userId, err := controller.refreshTokenService.FindUserIdbyRefreshToken(tokenId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.TokenNotFound)
			return
		}
		lib.ErrorResponse(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	token := controller.jWtService.GenerateToken(userId.Hex(), true, time.Minute*15)

	lib.JsonResponse(ctx, gin.H{
		"accessToken": token,
	})
}

// PUT /api/auth/logout/:tokenId
// Log the user out by removing the refresh token
func (controller *authController) Logout(ctx *gin.Context) {
	tokenId := ctx.Param("tokenId")

	err := controller.refreshTokenService.RemoveRefreshToken(tokenId)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, nil)
}

// POST /api/auth/forgot-pass
func (controller *authController) ForgotPass(ctx *gin.Context) {
	var dto dto.ForgotPass

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.UpdateActivationCode(*dto.Email)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.UserNotFound)
		return
	}

	err = controller.emailService.SendResetPassEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, nil)
}

// POST /api/auth/resend-activation-email
func (controller *authController) ResendActivationEmail(ctx *gin.Context) {
	var dto dto.ResendActivationEmail

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindUser(*dto.Email)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.UserNotFound)
		return
	}

	if user.Activated {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.UserAlreadyActivated)
		return
	}

	err = controller.emailService.SendActivationEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(ctx, nil)
}

// POST /api/auth/reset-password
func (controller *authController) ResetPass(ctx *gin.Context) {
	var dto dto.ResetPassword

	if err := ctx.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(ctx, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.ActivateUser(*dto.Email, *dto.Code, *dto.Password)
	if err != nil {
		lib.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(ctx, http.StatusUnprocessableEntity, errors.TokenExpired)
		return
	}

	lib.JsonResponse(ctx, nil)
}
