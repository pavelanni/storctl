package timeutil

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		created  time.Time
		expected string
	}{
		{
			name:     "just now",
			created:  time.Now(),
			expected: "0s",
		},
		{
			name:     "one hour ago",
			created:  time.Now().Add(-1 * time.Hour),
			expected: "1h",
		},
		{
			name:     "one day ago",
			created:  time.Now().Add(-24 * time.Hour),
			expected: "24h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAge(tt.created)
			if got != tt.expected {
				t.Errorf("FormatAge() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTtlToDuration(t *testing.T) {
	tests := []struct {
		name     string
		ttl      string
		wantErr  bool
		expected time.Duration
	}{
		{
			name:     "valid hours",
			ttl:      "24h",
			wantErr:  false,
			expected: 24 * time.Hour,
		},
		{
			name:     "valid minutes",
			ttl:      "30m",
			wantErr:  false,
			expected: 30 * time.Minute,
		},
		{
			name:    "invalid format",
			ttl:     "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			ttl:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TtlToDuration(tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("TtlToDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("TtlToDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseDeleteAfter(t *testing.T) {
	tests := []struct {
		name        string
		deleteAfter string
		wantZero    bool
	}{
		{
			name:        "valid timestamp",
			deleteAfter: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			wantZero:    false,
		},
		{
			name:        "invalid format",
			deleteAfter: "invalid",
			wantZero:    true,
		},
		{
			name:        "empty string",
			deleteAfter: "",
			wantZero:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDeleteAfter(tt.deleteAfter)
			if tt.wantZero && !got.IsZero() {
				t.Errorf("ParseDeleteAfter() = %v, want zero time", got)
			}
			if !tt.wantZero && got.IsZero() {
				t.Errorf("ParseDeleteAfter() returned zero time for valid input")
			}
		})
	}
}
