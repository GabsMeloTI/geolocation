package validation

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
)

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
	return re.ReplaceAllString(s, "")
}
