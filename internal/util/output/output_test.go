package output

import (
	"bytes"
	"testing"
)

type testStruct struct {
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Value  int               `json:"value" yaml:"value"`
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "simple struct",
			input: testStruct{
				Name:  "test",
				Value: 42,
			},
			expected: `{
  "name": "test",
  "value": 42
}
`,
			wantErr: false,
		},
		{
			name: "struct with labels",
			input: testStruct{
				Name: "test-with-labels",
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Value: 42,
			},
			expected: `{
  "name": "test-with-labels",
  "labels": {
    "key1": "value1",
    "key2": "value2"
  },
  "value": 42
}
`,
			wantErr: false,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "null\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := JSON(tt.input, buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.expected {
				t.Errorf("JSON() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "simple struct",
			input: testStruct{
				Name:  "test",
				Value: 42,
			},
			expected: `name: test
value: 42
`,
			wantErr: false,
		},
		{
			name: "struct with labels",
			input: testStruct{
				Name: "test-with-labels",
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Value: 42,
			},
			expected: `name: test-with-labels
labels:
  key1: value1
  key2: value2
value: 42
`,
			wantErr: false,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "null\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := YAML(tt.input, buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("YAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.expected {
				t.Errorf("YAML() = %v, want %v", got, tt.expected)
			}
		})
	}
}
