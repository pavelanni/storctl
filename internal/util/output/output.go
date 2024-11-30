package output

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

func JSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func YAML(v interface{}) error {
	// Convert the input to JSON bytes first (easy way to get a map)
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Create a map to hold the flattened structure
	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return err
	}

	// Extract and flatten typemeta and objectmeta
	if typeMeta, ok := data["typemeta"].(map[string]interface{}); ok {
		for k, v := range typeMeta {
			data[k] = v
		}
		delete(data, "typemeta")
	}
	if objectMeta, ok := data["objectmeta"].(map[string]interface{}); ok {
		for k, v := range objectMeta {
			data[k] = v
		}
		delete(data, "objectmeta")
	}

	return yaml.NewEncoder(os.Stdout).Encode(data)
}
