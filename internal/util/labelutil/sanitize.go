package labelutil

import "regexp"

var labelSanitizer = regexp.MustCompile(`^[_-]|[^a-zA-Z0-9_-]|[_-]$`)

// SanitizeValue ensures a label value contains only alphanumeric characters,
// dashes and underscores, and doesn't start or end with dash or underscore.
func SanitizeValue(value string) string {
	return labelSanitizer.ReplaceAllString(value, "")
}
