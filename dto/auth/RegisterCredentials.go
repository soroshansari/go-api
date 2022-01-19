package dto

//Register credential
type RegisterCredentials struct {
	Email     *string `json:"email" validate:"required,min=2,max=100"`
	Password  *string `json:"password" validate:"required,min=6,max=100"`
	Firstname *string `json:"firstname" validate:"required,min=2,max=100"`
	Lastname  *string `json:"lastname" validate:"required,min=2,max=100"`
}
