package validation

import (
	"github.com/go-playground/validator/v10"
	"net"
)

func Validate(data interface{}) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(data)
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
