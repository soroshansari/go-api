package server

import (
	"GoApp/controllers"
	"GoApp/db"
	"GoApp/providers"
	"log"
)

func Init() {
	var configs providers.Config = *providers.GetConfig()

	var dbClient = db.GetClient(configs)
	var userService db.UserService = db.NewUserService(dbClient, &configs)
	var refreshTokenService db.RefreshTokenService = db.NewRefreshTokenService(dbClient, &configs)
	var emailService providers.EmailService = providers.NewEmailService(&configs)
	var jwtService providers.JWTService = providers.NewJWTService(&configs)
	var healthController controllers.HealthController = controllers.HealthControllerHandler()
	var authController controllers.AuthController = controllers.AuthHandler(&jwtService, &userService, &refreshTokenService, &emailService, &configs)
	var userController controllers.UserController = controllers.UserHandler(&userService, &configs)

	r := NewRouter(&configs, &Controllers{
		healthController: healthController,
		authController:   authController,
		userController:   userController,
	}, &Providers{
		jwtService: jwtService,
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
