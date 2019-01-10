package api

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a short error code for failed API responses
type ErrorCode string

var (
	// ErrCodeNotFound describes an error locating a resource
	ErrCodeNotFound ErrorCode = "NOTFOUND"

	// ErrCodeBadBody describes an error reading a request body
	ErrCodeBadBody ErrorCode = "BADBODY"

	// ErrCodeBadAuth describes an error validating credentials
	ErrCodeBadAuth ErrorCode = "BADAUTH"

	// ErrCodeInternal describes an internal server error
	ErrCodeInternal ErrorCode = "INTERNALSERV"
)

// Error represents an API response error
type Error struct {
	Code    ErrorCode `json:"error"`
	Message string    `json:"message"`
	Status  int       `json:"-"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// NewError creates a new API response error
func NewError(code ErrorCode, message string, status int) Error {
	return Error{code, message, status}
}

// ErrNotFound generates a Not Found API error
func ErrNotFound(message string) Error {
	if message != "" {
		return NewError(ErrCodeNotFound, message, http.StatusNotFound)
	}
	return NewError(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
}

// ErrBadBody generates a Bad Request API error
func ErrBadBody(message string) Error {
	if message != "" {
		return NewError(ErrCodeBadBody, message, http.StatusBadRequest)
	}
	return NewError(ErrCodeBadBody, "Bad request body", http.StatusBadRequest)
}

// ErrInternal generates an Internal Server API error
func ErrInternal(message string) Error {
	if message != "" {
		return NewError(ErrCodeInternal, message, http.StatusInternalServerError)
	}
	return NewError(ErrCodeInternal, "Please try again later", http.StatusInternalServerError)
}

// ErrBadAuth generates an Unauthorized API error
func ErrBadAuth(message string) Error {
	if message != "" {
		return NewError(ErrCodeBadAuth, message, http.StatusUnauthorized)
	}
	return NewError(ErrCodeBadAuth, "Authentication failed", http.StatusUnauthorized)
}
