package dto

type UpdateUserDetails struct {
	Firstname *string `json:"firstname" validate:"required,min=2,max=100"`
	Lastname  *string `json:"lastname" validate:"required,min=2,max=100"`
}
