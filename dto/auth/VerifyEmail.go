package dto

//Login credential
type VerifyEmail struct {
	Email *string `json:"email" validate:"required,min=2,max=100"`
	Code  *string `json:"code" validate:"required,min=1,max=100"`
}
