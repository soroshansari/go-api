package controllers

import (
	"GoApp/db"
	dto "GoApp/dto/auth"
	"GoApp/lib"
	"GoApp/models"
	"GoApp/providers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//auth controllers interface
type AuthController interface {
	Register(c *gin.Context)
	VerifyEmail(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
	ForgotPass(c *gin.Context)
	ResendActivationEmail(c *gin.Context)
	ResetPass(c *gin.Context)
}

type authController struct {
	jWtService          providers.JWTService
	configs             providers.Config
	userService         db.UserService
	refreshTokenService db.RefreshTokenService
	emailService        providers.EmailService
	validate            validator.Validate
}

func AuthHandler(
	jWtService *providers.JWTService,
	userService *db.UserService,
	refreshTokenService *db.RefreshTokenService,
	emailService *providers.EmailService,
	configs *providers.Config,
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
func (controller *authController) Login(c *gin.Context) {
	var dto dto.LoginCredentials

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindUser(*dto.Email)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.IncorrectUserNameOrPassword)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*dto.Password)); err != nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.IncorrectUserNameOrPassword)
		return
	}

	if !user.Activated {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.UserNotVerified)
		return
	}

	token := controller.jWtService.GenerateToken(user.ID.Hex(), true, time.Minute*15)

	refreshToken, err := controller.refreshTokenService.CreateRefreshToken(user.ID)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, gin.H{
		"accessToken":  token,
		"refreshToken": refreshToken,
		"user":         models.GetUser(user, &controller.configs),
	})
}

// POST /api/auth/verify
func (controller *authController) VerifyEmail(c *gin.Context) {
	var dto dto.VerifyEmail

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.ActivateUser(*dto.Email, *dto.Code, "")
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.TokenExpired)
		return
	}

	lib.JsonResponse(c, nil)
}

// POST /api/auth/register
// Register a user
func (controller *authController) Register(c *gin.Context) {
	var dto dto.RegisterCredentials

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	isUserExists, err := controller.userService.UserExists(*dto.Email)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if isUserExists {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.UserExists)
		return
	}

	user, err := controller.userService.CreateUser(dto)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = controller.emailService.SendActivationEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

// PUT /api/auth/refresh/:tokenId
// Register a user
func (controller *authController) RefreshToken(c *gin.Context) {
	tokenId := c.Param("tokenId")

	userId, err := controller.refreshTokenService.FindUserIdbyRefreshToken(tokenId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.TokenNotFound)
			return
		}
		lib.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	token := controller.jWtService.GenerateToken(userId.Hex(), true, time.Minute*15)

	lib.JsonResponse(c, gin.H{
		"accessToken": token,
	})
}

// PUT /api/auth/logout/:tokenId
// Log the user out by removing the refresh token
func (controller *authController) Logout(c *gin.Context) {
	tokenId := c.Param("tokenId")

	err := controller.refreshTokenService.RemoveRefreshToken(tokenId)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

// POST /api/auth/forgot-pass
func (controller *authController) ForgotPass(c *gin.Context) {
	var dto dto.ForgotPass

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.UpdateActivationCode(*dto.Email)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.UserNotFound)
		return
	}

	err = controller.emailService.SendResetPassEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

// POST /api/auth/resend-activation-email
func (controller *authController) ResendActivationEmail(c *gin.Context) {
	var dto dto.ResendActivationEmail

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.FindUser(*dto.Email)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.UserNotFound)
		return
	}

	if user.Activated {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.UserAlreadyActivated)
		return
	}

	err = controller.emailService.SendActivationEmail(*user.Email, *user.FirstName, user.ActivationCode)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lib.JsonResponse(c, nil)
}

// POST /api/auth/reset-password
func (controller *authController) ResetPass(c *gin.Context) {
	var dto dto.ResetPassword

	if err := c.ShouldBind(&dto); err != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if validationErr := controller.validate.Struct(dto); validationErr != nil {
		lib.ErrorResponse(c, http.StatusBadRequest, validationErr.Error())
		return
	}

	user, err := controller.userService.ActivateUser(*dto.Email, *dto.Code, *dto.Password)
	if err != nil {
		lib.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		lib.ErrorResponse(c, http.StatusUnprocessableEntity, lib.TokenExpired)
		return
	}

	lib.JsonResponse(c, nil)
}
