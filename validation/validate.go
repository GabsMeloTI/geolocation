package validation

import (
	"github.com/go-playground/validator/v10"
)

func Validate(data interface{}) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(data)
}
