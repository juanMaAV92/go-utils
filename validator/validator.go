package validator

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/juanMaAV92/go-utils/errors"
)

// Validator wraps go-playground/validator with structured error formatting.
type Validator struct {
	v *validator.Validate
}

var (
	instance     *Validator
	instanceOnce sync.Once
)

// New returns the singleton Validator.
// JSON tag names are used in error messages instead of struct field names.
func New() *Validator {
	instanceOnce.Do(func() {
		v := validator.New()
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		instance = &Validator{v: v}
	})
	return instance
}

// Binder is implemented by any HTTP framework context that can decode a request body.
type Binder interface {
	Bind(any) error
}

// Validate validates a struct or a slice/array of structs.
// Returns *errors.ErrorResponse (HTTP 422) listing all failing fields.
func (val *Validator) Validate(data any) error {
	if data == nil {
		return nil
	}

	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if err := val.v.Var(rv.Interface(), "dive"); err != nil {
			return val.formatErrors(err)
		}
		return nil
	}

	if err := val.v.Struct(data); err != nil {
		return val.formatErrors(err)
	}
	return nil
}

// BindAndValidate binds the request body into req and then validates it.
func (val *Validator) BindAndValidate(binder Binder, req any) error {
	if err := binder.Bind(req); err != nil {
		return errors.New(http.StatusBadRequest, errors.InvalidRequestCode, []string{"invalid request format"})
	}
	return val.Validate(req)
}

func (val *Validator) formatErrors(err error) error {
	var messages []string
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			messages = append(messages, formatFieldError(fe))
		}
	}
	return errors.New(http.StatusUnprocessableEntity, errors.ValidationErrorCode, messages)
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Namespace()
	parts := strings.Split(field, ".")
	if len(parts) > 1 && !strings.HasPrefix(parts[0], "[") {
		field = strings.Join(parts[1:], ".")
	}

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "len":
		return field + " must be exactly " + fe.Param() + " characters"
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "numeric":
		return field + " must be numeric"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "datetime":
		return field + " must be a valid datetime"
	case "required_if":
		return field + " is required when " + fe.Param()
	case "required_unless":
		return field + " is required unless " + fe.Param()
	case "required_with":
		return field + " is required when " + fe.Param() + " is present"
	case "required_without":
		return field + " is required when " + fe.Param() + " is not present"
	default:
		return field + " is invalid (" + fe.Tag() + ")"
	}
}
