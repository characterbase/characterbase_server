package api

import (
	"cbs/models"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	validate "gopkg.in/go-playground/validator.v9"
)

// ContextKey represents a key for accessing API request context values
type ContextKey int

const (
	// UserContextKey represents a context key for accessing the session User from the request context
	UserContextKey ContextKey = iota

	// CollaboratorContextKey represents a context key for accessing the user collaborator from the request context
	CollaboratorContextKey
)

// Router represents a router with access to the database
type Router struct {
	*chi.Mux
	*Server
}

func addAPIHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")

	// Response status must be set after all other headers
	w.WriteHeader(status)
}

// SendError sends a failed API response to the ResponseWriter
func SendError(w http.ResponseWriter, err Error) {
	SendResponse(w, err, err.Status)
}

// SendResponse sends a successful API response to the ResponseWriter
func SendResponse(w http.ResponseWriter, data interface{}, status int) {
	serialized, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "An error occured.", http.StatusInternalServerError)
		return
	}
	addAPIHeaders(w, status)
	w.Write([]byte(serialized))
}

// ReadBody reads a JSON request body into a deserialized DTO
func ReadBody(body io.ReadCloser, out interface{}) error {
	data, err := ioutil.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

// ValidateDTO validates a request body DTO according to its struct tags
// TODO: Cache validator into a singleton?
func ValidateDTO(body interface{}) (*Error, error) {
	validator := validate.New()

	// Allow the validator to read struct JSON tags for error message generation
	validator.RegisterTagNameFunc(ValidatorJSON)

	err := validator.Struct(body)
	if err != nil {
		verr, err := models.GetValidationError(err)
		if err != nil {
			return nil, errors.New("failed to validate")
		}
		if verr != "" {
			return NewError(ErrCodeBadBody, verr, http.StatusBadRequest), nil
		}
	}
	return nil, nil
}

// ReadAndValidateBody combines ReadBody and ValidateDTO
func ReadAndValidateBody(body io.ReadCloser, out interface{}) (*Error, error) {
	if err := ReadBody(body, out); err != nil {
		return nil, err
	}
	valError, err := ValidateDTO(out)
	if err != nil {
		return nil, err
	}
	return valError, nil
}

// MustReadAndValidateBody combines ReadBody and ValidateDTO, enforcing strict validity
func MustReadAndValidateBody(w http.ResponseWriter, body io.ReadCloser, out interface{}) error {
	valError, err := ReadAndValidateBody(body, out)
	if err != nil {
		SendError(w, ErrBadBody(""))
		return err
	}
	if valError != nil {
		SendError(w, *valError)
		return errors.New("validation error")
	}
	return nil
}
