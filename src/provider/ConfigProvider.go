package provider

import "os"

type Configs struct {
	Port         string
	Env          string
	JwtSecret    string
	AppName      string
	MongoDbUrl   string
	DatabaseName string
}

func GetConfigs() *Configs {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return &Configs{
		Port:         port,
		Env:          env,
		JwtSecret:    os.Getenv("JWT_SECRET"),
		MongoDbUrl:   os.Getenv("MONGODB_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
	}
}
