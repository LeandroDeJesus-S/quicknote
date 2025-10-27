package validation

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

func ValidateStringNotEmpty(v string) (bool, string) {
	return strings.TrimSpace(v) != "", "field cannot be empty"
}
func ValidateRequired(v string) (bool, string) {
	return reflect.ValueOf(v).IsValid(), "field required"
}
func ValidateEmailPattern(v string) (bool, string) {
	matches, err := regexp.MatchString(
		"^[a-zA-Z0-9]+@[a-zA-Z_]+\\.[a-zA-Z]{2,3}$",
		v,
	)
	if err != nil {
		slog.Error("invalid regex pattern", "err", err)
	}

	return matches, "invalid email format"
}
func ValidateMinMaxLen(min, max int) ValidatorFunc {
	return func(v string) (bool, string) {
		size := utf8.RuneCountInString(v)
		if size < min || size > max {
			return false, fmt.Sprintf("the value must have between %d and %d characters", min, max)
		}
		return true, ""
	}
}
