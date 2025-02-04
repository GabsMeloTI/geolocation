package validation

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// ParseStringToInt64 converts a string representing a UserID to an integer (int64).
// Returns 0 and nil if the input string is empty.
// Returns an error if the conversion fails.
func ParseStringToInt64(strUserID string) (int64, error) {
	if strUserID == "" {
		return 0, nil
	}
	userID, err := strconv.ParseInt(strUserID, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func ParseStringToInt32(strUserID string) (int32, error) {
	if strUserID == "" {
		return 0, nil
	}
	userID, err := strconv.ParseInt(strUserID, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(userID), nil
}

func ExtractCNPJNumbers(strDocument string) string {
	var numbers strings.Builder
	for _, char := range strDocument {
		if unicode.IsDigit(char) {
			numbers.WriteRune(char)
		}
	}
	return numbers.String()
}

func GetFirstDigits(strDocument string, number int) (string, error) {
	if len(strDocument) > number {
		return strDocument[:number], nil
	}
	return strDocument, errors.New("invalid Document")
}
