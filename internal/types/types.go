package types

type Student struct {
	Name string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age int `json:"age" validate:"required,numeric,min=18"`
}