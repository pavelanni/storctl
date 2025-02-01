// Package output contains the functions to format and write data to the console.
// It includes the functions to format data as JSON and YAML.
package output

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// JSON formats data as JSON and writes it to the specified writer
func JSON(data interface{}, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// YAML formats data as YAML and writes it to the specified writer
func YAML(data interface{}, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}
