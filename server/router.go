package server

import (
	"GoApp/controllers"
	"GoApp/middlewares"
	"GoApp/providers"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type Controllers struct {
	healthController controllers.HealthController
	authController   controllers.AuthController
	userController   controllers.UserController
}

type Providers struct {
	jwtService providers.JWTService
}

func NewRouter(configs *providers.Config, controllers *Controllers, providers *Providers) *gin.Engine {
	router := gin.New()

	config := cors.DefaultConfig()
	if configs.AllowOrigin != "" {
		config.AllowOrigins = []string{configs.AllowOrigin}
	} else {
		config.AllowAllOrigins = true
	}
	config.AllowMethods = []string{"*"}
	config.AllowHeaders = []string{"*"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// Global middlewares
	// Logger middlewares will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())

	// Recovery middlewares recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	router.Use(static.Serve("/public", static.LocalFile("public", false)))

	// Routes
	router.GET("/", controllers.healthController.Status)

	v1 := router.Group("v1")
	v1.Use(middlewares.AuthMiddleware(configs.AuthKey))
	{
		auth := v1.Group("auth")
		{
			auth.POST("login", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "login"), controllers.authController.Login)
			auth.POST("register", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "register"), controllers.authController.Register)
			auth.POST("verify", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "verify"), controllers.authController.VerifyEmail)
			auth.POST("forgot-password", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "forgot-password"), controllers.authController.ForgotPass)
			auth.POST("resend-activation-email", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "resend-activation-email"), controllers.authController.ResendActivationEmail)
			auth.POST("reset-password", middlewares.RecaptchaMiddleware(configs.RecaptchaSecret, "reset-password"), controllers.authController.ResetPass)
			auth.PUT("refresh/:tokenId", controllers.authController.RefreshToken)
			auth.PUT("logout/:tokenId", controllers.authController.Logout)
		}

		user := v1.Group("user")
		user.Use(middlewares.AuthorizeJWT(providers.jwtService))
		{
			user.GET("details", controllers.userController.Me)
			user.POST("change-password", controllers.userController.ChangePassword)
			user.POST("profile", controllers.userController.UploadProfile)
			user.POST("details", controllers.userController.UpdateUserDetails)
		}
	}
	return router

}
