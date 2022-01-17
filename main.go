package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
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
	var jwtService provider.JWTService = provider.JWTAuthService(&configs)
	var authController controller.AuthController = controller.AuthHandler(&jwtService, &userService, &configs)
	var userController controller.UserController = controller.UserHandler(&userService, &configs)

	app := gin.New()

	if configs.Env == "local" {
		fmt.Println(configs.Env)
		config := cors.DefaultConfig()
		// config.AllowOrigins = []string{"http://google.com"}
		// config.AllowOrigins == []string{"http://google.com", "http://facebook.com"}
		config.AllowAllOrigins = true
		config.AllowMethods = []string{"*"}
		config.AllowHeaders = []string{"*"}
		config.AllowCredentials = true

		app.Use(cors.New(config))
	}

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())

	// app.Use(static.Serve("/", static.LocalFile("/client", false)))

	// books := app.Group("/api/books")
	// books.Use(middleware.AuthorizeJWT())
	// {
	// 	books.GET("/", controller.FindBooks)
	// 	books.POST("/", controller.CreateBook)
	// 	books.GET("/:id", controller.FindBook)
	// 	books.PATCH("/:id", controller.UpdateBook)
	// 	books.DELETE("/:id", controller.DeleteBook)
	// }

	auth := app.Group("/api/auth")
	{
		auth.POST("login", authController.Login)
		auth.POST("register", authController.Register)
		auth.POST("refresh-token", authController.RefreshToken)
		// auth.PUT("logout", authController.Logout)
	}
	user := app.Group("/api/user")
	user.Use(middleware.AuthorizeJWT(&configs))
	{
		user.GET("/detail", userController.Me)
	}

	app.Run()
}
