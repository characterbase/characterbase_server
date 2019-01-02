package dtos

// ReqLogIn represents a request DTO for accessing a User account session
type ReqLogIn struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}
