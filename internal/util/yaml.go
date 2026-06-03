package util

import (
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

func ReadYAMLKeyValue(fpath string, key string) (string, error) {
	data := make(map[string]any)

	content, err := os.ReadFile(fpath)
	if err != nil {
		return "", nil
	}

	if err := yaml.Unmarshal(content, &data); err != nil {
		return "", err
	}

	return GetNestedMapValue(data, key), nil
}

func WriteYAMLKeyValue(fpath string, key, value string) error {
	cfg := make(map[string]any)

	if content, err := os.ReadFile(fpath); err == nil {
		if err := yaml.Unmarshal(content, &cfg); err != nil {
			return err
		}
	}

	SetNestedMapValue(cfg, key, value)

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	return os.WriteFile(fpath, []byte(buf.String()), 0644)
}
