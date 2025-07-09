package zcfg

import (
	"encoding/json"
	"strings"
)

// Config represents the configuration loader
type Config struct {
	options *Options
	data    map[string]any
}

// New creates a new Config instance with the given options
func New(opts ...Option) *Config {
	c := &Config{
		options: &Options{
			IgnoreCase:  true,
			TagName:     "meta",
			UseDefaults: true,
			UseEnv:      true,
		},
		data: make(map[string]any),
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			if opt != nil {
				opt(c)
			}
		}
	}

	return c
}

func Must[T any](filename string, opts ...Option) *T {
	var v T
	if err := Load(filename, &v, opts...); err != nil {
		panic(err)
	}
	return &v
}

// Load is a global function to create a new Config and load configuration from a file
func Load(filename string, v any, opts ...Option) error {
	c := New(opts...)

	if err := c.LoadFile(filename); err != nil {
		return err
	}

	if c.options.IgnoreCase {
		c.buildCaseInsensitiveCache()
	}

	return c.Parse(v)
}

// LoadFromJsonBytes loads config into v from content json bytes.
func LoadFromJsonBytes(content []byte, v any, opts ...Option) error {
	c := New(opts...)

	if err := json.Unmarshal(content, &c.data); err != nil {
		return err
	}

	if c.options.IgnoreCase {
		c.buildCaseInsensitiveCache()
	}

	return c.Parse(v)
}

// buildCaseInsensitiveCache rebuilds the case-insensitive cache
func (c *Config) buildCaseInsensitiveCache() {
	c.data = buildCaseInsensitiveMap(c.data)
}

func buildCaseInsensitiveMap(data map[string]any) map[string]any {
	caseInsensitive := make(map[string]any)
	for key, val := range data {
		lowerKey := strings.ToLower(key)
		switch v := val.(type) {
		case map[string]any:
			caseInsensitive[lowerKey] = buildCaseInsensitiveMap(v)
		default:
			caseInsensitive[lowerKey] = v
		}
	}
	return caseInsensitive
}
