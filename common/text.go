package common

import (
	"regexp"
	"strings"
)

// RemoveNonNumeric takes a string and removes all non-numeric characters (anything that is not a digit).
// It uses a regular expression to perform the removal and returns the modified string with only numeric characters.
func RemoveNonNumeric(str string) string {
	regex := regexp.MustCompile(`[^0-9]+`)
	return regex.ReplaceAllString(str, "")
}

// RemovePrefix removes a specified prefix from a string.
// It checks if the input string starts with the given prefix.
// If it does, the prefix is removed, and the result is returned.
// If not, the original string is returned unchanged.
func RemovePrefix(prefix, input string) string {
	if strings.HasPrefix(input, prefix) {
		return strings.TrimPrefix(input, prefix)
	} else {
		return input
	}
}
