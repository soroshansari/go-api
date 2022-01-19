package dto

//Login credential
type ForgotPass struct {
	Email *string `json:"email" validate:"required,min=2,max=100"`
}
