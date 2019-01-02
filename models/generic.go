package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
)

// Generic represents a generic database model with ID support
type Generic struct {
	ID string `json:"id" gorm:"primary_key"`
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
	if !ok {
		return "", errors.New("failed to validate")
	}
	if len(valError) > 0 {
		return fmt.Sprintf("%v failed validation", valError[0].Field()), nil
	}
	return "", nil
}

// BeforeCreate generates a new ID for the model
func (g *Generic) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", ksuid.New().String())
	return nil
}
