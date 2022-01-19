package dto

//Login credential
type ChangePassword struct {
	OldPassword *string `json:"oldpassword" validate:"required,min=1,max=100"`
	NewPassword *string `json:"newPassword" validate:"required,min=6,max=100"`
}
