package validation

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
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

func ParseNullStringToFloat(nullString sql.NullString) (float64, error) {
	if nullString.Valid {
		return strconv.ParseFloat(nullString.String, 64)
	}
	return 0, fmt.Errorf("valor nulo")
}
func ParseStringToFloat(text string) (float64, error) {
	return strconv.ParseFloat(text, 64)
}

func GetStringFromNull(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return ""
}

func RemoveHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, " ")
}
