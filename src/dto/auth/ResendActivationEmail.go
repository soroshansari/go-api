package dto

//Login credential
type ResendActivationEmail struct {
	Email *string `json:"email" validate:"required,min=2,max=100"`
}
