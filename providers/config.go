package providers

import "os"

type Config struct {
	Port            string
	Env             string
	JwtSecret       string
	AppName         string
	MongoDbUrl      string
	DatabaseName    string
	SmtpSender      string
	SmtpHost        string
	SmtpPort        string
	SmtpPassword    string
	VerifyUrl       string
	ResetPassUrl    string
	RecaptchaSecret string
	AllowOrigin     string
	Domain          string
	AuthKey         string
}

func GetConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return &Config{
		Port:            port,
		Env:             env,
		JwtSecret:       os.Getenv("JWT_SECRET"),
		MongoDbUrl:      os.Getenv("MONGODB_URL"),
		DatabaseName:    os.Getenv("DATABASE_NAME"),
		SmtpSender:      os.Getenv("SMTP_SENDER"),
		SmtpHost:        os.Getenv("SMTP_HOST"),
		SmtpPort:        os.Getenv("SMTP_PORT"),
		SmtpPassword:    os.Getenv("SMTP_PASSWORD"),
		VerifyUrl:       os.Getenv("FE_VERIFY_URL"),
		ResetPassUrl:    os.Getenv("FE_RESET_PASS_URL"),
		RecaptchaSecret: os.Getenv("RECAPTCHA_SECRET"),
		AllowOrigin:     os.Getenv("ALLOWED_ORIGIN"),
		Domain:          os.Getenv("DOMAIN"),
		AuthKey:         os.Getenv("AUTH_KEY"),
	}
}
