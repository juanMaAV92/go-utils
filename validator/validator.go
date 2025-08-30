package validator

import (
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/juanMaAV92/go-utils/errors"
)

type Validator struct {
	validator *validator.Validate
}

var (
	singletonValidator     *Validator
	singletonValidatorOnce sync.Once
)

func New() *Validator {
	singletonValidatorOnce.Do(func() {
		singletonValidator = &Validator{
			validator: validator.New(),
		}
	})
	return singletonValidator
}

func GetInstance() *Validator {
	return singletonValidator
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

func (v *Validator) formatValidationErrors(err error) error {
	var validationErrors []string

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			validationErrors = append(validationErrors, err.Field()+" is required")
		case "email":
			validationErrors = append(validationErrors, "Invalid email format")
		case "min":
			if err.Field() == "Password" {
				validationErrors = append(validationErrors, "Password must be at least "+err.Param()+" characters long")
			} else {
				validationErrors = append(validationErrors, err.Field()+" must be at least "+err.Param()+" characters")
			}
		case "max":
			validationErrors = append(validationErrors, err.Field()+" must be at most "+err.Param()+" characters")
		case "uuid":
			validationErrors = append(validationErrors, err.Field()+" must be a valid UUID")
		case "oneof":
			validationErrors = append(validationErrors, err.Field()+" must be one of: "+err.Param())
		case "gte":
			validationErrors = append(validationErrors, err.Field()+" must be greater than or equal to "+err.Param())
		case "lte":
			validationErrors = append(validationErrors, err.Field()+" must be less than or equal to "+err.Param())
		case "numeric":
			validationErrors = append(validationErrors, err.Field()+" must be numeric")
		case "alpha":
			validationErrors = append(validationErrors, err.Field()+" must contain only letters")
		case "alphanum":
			validationErrors = append(validationErrors, err.Field()+" must contain only letters and numbers")
		case "len":
			validationErrors = append(validationErrors, err.Field()+" must be exactly "+err.Param()+" characters")
		case "gt":
			validationErrors = append(validationErrors, err.Field()+" must be greater than "+err.Param())
		case "lt":
			validationErrors = append(validationErrors, err.Field()+" must be less than "+err.Param())
		case "url":
			validationErrors = append(validationErrors, err.Field()+" must be a valid URL")
		case "datetime":
			validationErrors = append(validationErrors, err.Field()+" must be a valid datetime format")
		default:
			validationErrors = append(validationErrors, err.Field()+" is invalid")
		}
	}

	return errors.New(http.StatusBadRequest, errors.ValidationErrorCode, validationErrors)
}

func BindAndValidate(c interface{}, req interface{}) error {
	if binder, ok := c.(interface{ Bind(interface{}) error }); ok {
		if err := binder.Bind(req); err != nil {
			return errors.New(http.StatusBadRequest, errors.InvalidRequestCode, []string{"Invalid request format"})
		}
	}

	return GetInstance().Validate(req)
}
