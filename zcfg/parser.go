package zcfg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// parseConfigFile parses configuration file and returns rawMap
func parseConfigFile(filename string) (map[string]any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return parseJSON(data)
	case ".yaml", ".yml":
		return parseYAML(data)
	case ".toml":
		return parseTOML(data)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// parseJSON parses JSON data and returns rawMap
func parseJSON(data []byte) (map[string]any, error) {
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return rawMap, nil
}

// parseYAML parses YAML data and returns rawMap
func parseYAML(data []byte) (map[string]any, error) {
	var rawMap map[string]any
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return rawMap, nil
}

// parseTOML parses TOML data and returns rawMap
func parseTOML(data []byte) (map[string]any, error) {
	var rawMap map[string]any
	if err := toml.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return rawMap, nil
}

// parseJSONBytes parses JSON bytes and returns rawMap
func parseJSONBytes(data []byte) (map[string]any, error) {
	return parseJSON(data)
}
