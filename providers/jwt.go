package providers

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//jwt service
type JWTService interface {
	GenerateToken(userId string, isUser bool, expiresIn time.Duration) string
	ValidateToken(token string) (*jwt.Token, error)
}
type authCustomClaims struct {
	UserId string `json:"sub"`
	User   bool   `json:"user"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issure    string
}

//auth-jwt
func NewJWTService(configs *Config) JWTService {
	return &jwtServices{
		secretKey: configs.JwtSecret,
		issure:    configs.AppName,
	}
}

func (service *jwtServices) GenerateToken(userId string, isUser bool, expiresIn time.Duration) string {

	claims := &authCustomClaims{
		userId,
		isUser,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresIn).Unix(),
			Issuer:    service.issure,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		panic(err)
	}
	return t
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token: %s", token.Header["alg"])

		}
		return []byte(service.secretKey), nil
	})

}
