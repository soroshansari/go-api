package dto

//Register credential
type RegisterCredentials struct {
	Email     *string `json:"email" validate:"required,min=2,max=100"`
	Password  *string `json:"password" validate:"required,min=6,max=100"`
	FirstName *string `json:"firstName" validate:"required,min=2,max=100"`
	LastName  *string `json:"lastName" validate:"required,min=2,max=100"`
}
