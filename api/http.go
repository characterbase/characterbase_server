package api

import (
	"cbs/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"
)

// ContextKey represents a key for accessing API request context values
type ContextKey int

const (
	// UserContextKey represents a context key for accessing the session User from the request context
	UserContextKey ContextKey = iota

	// CollaboratorContextKey represents a context key for accessing the user collaborator from the request context
	CollaboratorContextKey

	// UniverseContextKey represents a context key for accessing the universe from the request context
	UniverseContextKey

	// CharacterContextKey represents a context key for accessing the character from the request context
	CharacterContextKey
)

// Router represents a router with access to the database
type Router struct {
	*chi.Mux
	*Server
}

// Handler represents a generic HTTP handler with improved error-handling support
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		apierr, ok := err.(Error)
		if !ok {
			SendError(w, ErrInternal(err.Error()))
		} else {
			SendError(w, apierr)
		}
	}
}

func addAPIHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")

	// Response status must be set after all other headers
	w.WriteHeader(status)
}

// ValidatorJSONPlugin serves as a plugin for go-validator
// that recognizes JSON tags when validating structs
func ValidatorJSONPlugin(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

// GetValidationError returns a simplified error
// from go-validator validation results
func GetValidationError(err error) (string, error) {
	valError, ok := err.(validator.ValidationErrors)
	fmt.Printf("%T\n", err)
	if !ok {
		return "", errors.New("failed to validate")
	}
	if len(valError) > 0 {
		fmt.Println(valError[0])
		return fmt.Sprintf("%v failed validation", valError[0].Field()), nil
	}
	return "", nil
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
		return ErrBadBody("Failed to parse body")
	}
	return nil
}

// ValidateDTO validates a request body DTO according to its struct tags
// TODO: Cache validator into a singleton?
func ValidateDTO(body interface{}) (*Error, error) {
	v := validator.New()

	// Allow the validator to read struct JSON tags for error message generation
	v.RegisterTagNameFunc(ValidatorJSONPlugin)

	// Allow the validator to handle special validation cases pertaining to universe guide fields
	v.RegisterStructValidation(UniverseFieldStructLevelValidation, models.UniverseGuideField{})

	err := v.Struct(body)
	if err != nil {
		verr, err := GetValidationError(err)
		if err != nil {
			return nil, err
		}
		if verr != "" {
			err := ErrBadBody(verr)
			return &err, nil
		}
	}
	return nil, nil
}

// ReadAndValidateBody combines ReadBody and ValidateDTO
func ReadAndValidateBody(body io.ReadCloser, out interface{}) error {
	if err := ReadBody(body, out); err != nil {
		return err
	}
	valError, err := ValidateDTO(out)
	if err != nil {
		return err
	}
	if valError != nil {
		return *valError
	}
	return nil
}
