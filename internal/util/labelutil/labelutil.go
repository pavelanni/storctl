// Package labelutil contains the functions to sanitize and merge labels.
package labelutil

import "regexp"

var labelSanitizer = regexp.MustCompile(`^[_-]|[^a-zA-Z0-9_-]|[_-]$`)

// SanitizeValue ensures a label value contains only alphanumeric characters,
// dashes and underscores, and doesn't start or end with dash or underscore.
func SanitizeValue(value string) string {
	return labelSanitizer.ReplaceAllString(value, "")
}

// MergeLabels combines two maps of labels, with override taking precedence over base.
// If a key exists in both maps, the value from override is used.
func MergeLabels(base, override map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy all base labels
	for k, v := range base {
		result[k] = v
	}

	// Apply overrides
	for k, v := range override {
		result[k] = v
	}

	return result
}
