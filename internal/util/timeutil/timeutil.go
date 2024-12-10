package timeutil

import (
	"fmt"
	"time"
)

// ParseDeleteAfter parses a date string in the format "2006-01-02-15-04"
// Returns zero time if parsing fails
func ParseDeleteAfter(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02-15-04", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// FormatAge converts a timestamp into a human-readable duration string
// Examples: "37s", "3m12s", "14h", "2d4h"
func FormatAge(t time.Time) string {
	duration := time.Since(t)

	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := int(hours / 24)

	if days > 0 {
		remainingHours := hours % 24
		if remainingHours > 0 {
			return fmt.Sprintf("%dd%dh", days, remainingHours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		remainingSeconds := seconds % 60
		if remainingSeconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", seconds)
}

func TtlToDuration(ttl string) (time.Duration, error) {
	if ttl == "" {
		return 0, fmt.Errorf("ttl cannot be empty")
	}

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return 0, fmt.Errorf("invalid ttl format: %w", err)
	}

	return duration, nil
}

func FormatDeleteAfter(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02-15-04")
}
