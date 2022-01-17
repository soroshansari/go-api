package dto

//Login credential
type LoginCredentials struct {
	Email    *string `json:"email" validate:"required,min=2,max=100"`
	Password *string `json:"password" validate:"required,min=1,max=100"`
}
