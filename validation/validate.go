package validation

import (
	"github.com/go-playground/validator/v10"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func Validate(data interface{}) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(data)
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasDigit, hasSpecial bool

	for _, c := range password {
		switch {

		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	specialCharRegex := regexp.MustCompile(`[!@#$%^&*()\-_=+\[\]{}|;:'",.<>?/\\` + "`~]")

	hasSpecial = specialCharRegex.MatchString(password)

	if !hasUpper || !hasDigit || !hasSpecial {
		return false
	}

	return true
}

func ValidateCPF(cpf string) bool {
	reg := regexp.MustCompile(`\D`)
	cpf = reg.ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	for i := 0; i < 10; i++ {
		if cpf == strings.Repeat(strconv.Itoa(i), 11) {
			return false
		}
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(cpf[i]))
		if err != nil {
			return false
		}
		sum += digit * (10 - i)
	}
	var firstCheck int
	remainder := sum % 11
	if remainder < 2 {
		firstCheck = 0
	} else {
		firstCheck = 11 - remainder
	}
	if firstCheck != int(cpf[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		digit, err := strconv.Atoi(string(cpf[i]))
		if err != nil {
			return false
		}
		sum += digit * (11 - i)
	}
	var secondCheck int
	remainder = sum % 11
	if remainder < 2 {
		secondCheck = 0
	} else {
		secondCheck = 11 - remainder
	}
	return secondCheck == int(cpf[10]-'0')
}

func ValidatePhone(phone string) bool {
	re := regexp.MustCompile(`^(?:\+55\s?)?(?:\(?\d{2}\)?\s?)?(?:9\d{4}|\d{4})-?\d{4}$`)
	return re.MatchString(phone)
}

func ValidateCNH(cnh string) bool {
	reg := regexp.MustCompile(`\D`)
	cnh = reg.ReplaceAllString(cnh, "")

	if len(cnh) != 11 {
		return false
	}

	sum1 := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(cnh[i]))
		if err != nil {
			return false
		}
		sum1 += digit * (9 - i)
	}
	var check1 int
	remainder1 := sum1 % 11
	if remainder1 == 10 {
		check1 = 0
	} else {
		check1 = remainder1
	}

	sum2 := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(cnh[i]))
		if err != nil {
			return false
		}
		sum2 += digit * (1 + i)
	}
	sum2 += check1 * 9
	var check2 int
	remainder2 := sum2 % 11
	if remainder2 == 10 {
		check2 = 0
	} else {
		check2 = remainder2
	}

	digit10, err := strconv.Atoi(string(cnh[9]))
	if err != nil {
		return false
	}
	digit11, err := strconv.Atoi(string(cnh[10]))
	if err != nil {
		return false
	}

	return check1 == digit10 && check2 == digit11
}
