package labelutil

import "testing"

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "test-label",
			expected: "test-label",
		},
		{
			name:     "uppercase letters",
			input:    "TEST",
			expected: "TEST",
		},
		{
			name:     "spaces",
			input:    "test label",
			expected: "testlabel",
		},
		{
			name:     "special characters",
			input:    "test@label#123",
			expected: "testlabel123",
		},
		{
			name:     "starts with dash",
			input:    "-test-label",
			expected: "test-label",
		},
		{
			name:     "ends with underscore",
			input:    "test_label_",
			expected: "test_label",
		},
		{
			name:     "multiple special chars",
			input:    "$$test##label%%",
			expected: "testlabel",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeValue(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeLabels(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name: "merge with override",
			base: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			override: map[string]string{
				"key2": "new-value2",
				"key3": "value3",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "new-value2",
				"key3": "value3",
			},
		},
		{
			name: "empty override",
			base: map[string]string{
				"key1": "value1",
			},
			override: map[string]string{},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:     "empty base",
			base:     map[string]string{},
			override: map[string]string{"key1": "value1"},
			expected: map[string]string{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeLabels(tt.base, tt.override)
			if len(got) != len(tt.expected) {
				t.Errorf("MergeLabels() returned map of size %d, want %d", len(got), len(tt.expected))
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("MergeLabels()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
