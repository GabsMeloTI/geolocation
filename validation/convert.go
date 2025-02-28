package validation

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func FormatActiveDuration(start time.Time) string {
	now := time.Now()
	years := now.Year() - start.Year()
	months := int(now.Month()) - int(start.Month())

	if now.Day() < start.Day() {
		months--
	}
	if months < 0 {
		years--
		months += 12
	}

	var parts []string
	if years > 0 {
		if years == 1 {
			parts = append(parts, "1 ano")
		} else {
			parts = append(parts, fmt.Sprintf("%d anos", years))
		}
	}
	if months > 0 {
		if months == 1 {
			parts = append(parts, "1 mês")
		} else {
			parts = append(parts, fmt.Sprintf("%d meses", months))
		}
	}
	if len(parts) == 0 {
		return "Ativo a: menos de um mês"
	}
	return "Ativo a: " + strings.Join(parts, " e ")
}
