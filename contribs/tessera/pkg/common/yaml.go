package common

import (
	"os"

	"github.com/goccy/go-yaml"
)

// LoadYAML loads the given YAML file
func LoadYAML[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var f T
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	return &f, nil
}
