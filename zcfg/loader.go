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

// LoadFile loads configuration from a file
func (c *Config) LoadFile(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	switch ext {
	case ".json":
		return c.LoadJSON(data)
	case ".yaml", ".yml":
		return c.LoadYAML(data)
	case ".toml":
		return c.LoadTOML(data)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

// LoadJSON loads configuration from JSON data
func (c *Config) LoadJSON(data []byte) error {
	return json.Unmarshal(data, &c.data)
}

// LoadYAML loads configuration from YAML data
func (c *Config) LoadYAML(data []byte) error {
	return yaml.Unmarshal(data, &c.data)
}

// LoadTOML loads configuration from TOML data
func (c *Config) LoadTOML(data []byte) error {
	return toml.Unmarshal(data, &c.data)
}
