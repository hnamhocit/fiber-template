// Package validator provides a singleton struct validator used by handlers.
// Tags: required, email, min, max, oneof, uuid, gte, lte... (see go-playground/validator docs).
// Custom validations can be registered via RegisterCustom() at app startup.
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func init() {
	// Use JSON tag names (not Go field names) in validation error messages,
	// so the client sees "email" not "Email".
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

// FieldError is one per-field validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate validates a struct using its `validate` tags.
// Returns nil when valid, or a slice of FieldError when invalid.
func Validate(s any) []FieldError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var out []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		out = append(out, FieldError{
			Field:   e.Field(), // JSON tag thanks to RegisterTagNameFunc
			Message: humanMessage(e),
		})
	}
	return out
}

// humanMessage turns a validator error into a friendly English message.
// Extend this switch as the project grows new custom rules.
func humanMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "uuid":
		return "must be a valid UUID"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", e.Param())
	case "gte":
		return fmt.Sprintf("must be >= %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be <= %s", e.Param())
	case "gt":
		return fmt.Sprintf("must be > %s", e.Param())
	case "lt":
		return fmt.Sprintf("must be < %s", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "alphanum":
		return "must be alphanumeric"
	case "alpha":
		return "must be alphabetic"
	case "numeric":
		return "must be numeric"
	default:
		return fmt.Sprintf("failed %s validation (param=%s)", e.Tag(), e.Param())
	}
}
