package validation

import "strings"

func ValidateStringNotEmpty(v string) (bool, string) {
	return strings.TrimSpace(v) != "", "field required"
}
