package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"GoApp/src/controller"
	"GoApp/src/database"
	"GoApp/src/middleware"
	"GoApp/src/provider"
	"GoApp/src/service"
)

func main() {

	var configs provider.Configs = *provider.GetConfigs()

	var dbClient = database.GetClient(configs)
	var userService service.UserService = service.StaticUserService(dbClient, &configs)
	var refreshTokenService service.RefreshTokenService = service.StaticRefreshTokenService(dbClient, &configs)
	var emailService provider.EmailService = provider.StaticEmailService(&configs)
	var jwtService provider.JWTService = provider.JWTAuthService(&configs)
	var mainController controller.MainController = controller.MainControllerHandler()
	var authController controller.AuthController = controller.AuthHandler(&jwtService, &userService, &refreshTokenService, &emailService, &configs)
	var userController controller.UserController = controller.UserHandler(&userService, &configs)

	app := gin.New()

	config := cors.DefaultConfig()
	if configs.AllowOrigin != "" {
		config.AllowOrigins = []string{configs.AllowOrigin}
	} else {
		config.AllowAllOrigins = true
	}
	config.AllowMethods = []string{"*"}
	config.AllowHeaders = []string{"*"}
	config.AllowCredentials = true

	app.Use(cors.New(config))

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())

	app.Use(static.Serve("/public", static.LocalFile("public", false)))

	// Routes
	app.GET("/", mainController.HealthCheck)
	auth := app.Group("/api/auth")
	{
		auth.POST("login", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "login"), authController.Login)
		auth.POST("register", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "register"), authController.Register)
		auth.POST("verify", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "verify"), authController.VerifyEmail)
		auth.POST("forgot-password", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "forgot-password"), authController.ForgotPass)
		auth.POST("resend-activation-email", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "resend-activation-email"), authController.ResendActivationEmail)
		auth.POST("reset-password", middleware.RecaptchaMiddleware(configs.RecaptchaSecret, "reset-password"), authController.ResetPass)
		auth.PUT("refresh/:tokenId", authController.RefreshToken)
		auth.PUT("logout/:tokenId", authController.Logout)
	}
	user := app.Group("/api/user")
	user.Use(middleware.AuthorizeJWT(&configs))
	{
		user.GET("detail", userController.Me)
		user.POST("change-password", userController.ChangePassword)
		user.POST("profile", userController.UploadProfile)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
