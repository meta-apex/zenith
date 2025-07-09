package zcfg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// tagOptions represents the parsed tag options
type tagOptions struct {
	hasDefault   bool
	defaultValue any
	hasRange     bool
	rangeMin     int64
	rangeMax     int64
	hasOptions   bool
	options      []string
	optional     bool
}

// parseTag parses a struct tag
func parseTag(tag string) (string, tagOptions) {
	if tag == "" {
		return "", tagOptions{}
	}

	parts := strings.Split(tag, ",")
	name := parts[0]

	opts := tagOptions{}
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch {
		case strings.HasPrefix(part, "default="):
			opts.hasDefault = true
			opts.defaultValue = parseDefaultValue(part[8:])

		case strings.HasPrefix(part, "range="):
			rMin, rMax, err := parseRange(part[6:])
			if err == nil {
				opts.hasRange = true
				opts.rangeMin = rMin
				opts.rangeMax = rMax
			}

		case strings.HasPrefix(part, "options="):
			opts.hasOptions = true
			opts.options = strings.Split(part[8:], "|")

		case part == "optional":
			opts.optional = true
		}
	}

	return name, opts
}

// parseDefaultValue parses a default value from a tag
func parseDefaultValue(val string) any {
	// Handle array default values
	if strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}") {
		val = strings.TrimPrefix(strings.TrimSuffix(val, "}"), "{")
		return strings.Split(val, ";")
	}

	// Handle numeric values
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}

	// Handle boolean values
	if b, err := strconv.ParseBool(val); err == nil {
		return b
	}

	// Return as string by default
	return val
}

// parseRange parses a range constraint from a tag
func parseRange(val string) (min int64, max int64, err error) {
	if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	val = strings.TrimPrefix(strings.TrimSuffix(val, "]"), "[")
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	min, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range minimum")
	}

	max, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range maximum")
	}

	if min > max {
		return 0, 0, fmt.Errorf("range minimum cannot be greater than maximum")
	}

	return min, max, nil
}

// validateRange validates a numeric value against a range constraint
func validateRange(val int64, opts tagOptions) error {
	if !opts.hasRange {
		return nil
	}

	if val < opts.rangeMin || val > opts.rangeMax {
		return fmt.Errorf("value must be between %d and %d", opts.rangeMin, opts.rangeMax)
	}

	return nil
}

// validateOptions validates a string value against allowed options
func validateOptions(val string, opts tagOptions) error {
	if !opts.hasOptions {
		return nil
	}

	for _, opt := range opts.options {
		if val == opt {
			return nil
		}
	}

	return fmt.Errorf("value must be one of: %s", strings.Join(opts.options, ", "))
}

// resolveEnvVars resolves environment variables in a value
func (c *Config) resolveEnvVars(val any) any {
	if !c.options.UseEnv {
		return val
	}

	switch v := val.(type) {
	case string:
		return resolveEnvVarsInString(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = c.resolveEnvVars(item)
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for k, item := range v {
			result[k] = c.resolveEnvVars(item)
		}
		return result
	default:
		return val
	}
}

// resolveEnvVarsInString resolves environment variables in a string
func resolveEnvVarsInString(s string) string {
	result := s
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		envVar := result[start+2 : end]
		defaultVal := ""

		// Handle default value
		if idx := strings.Index(envVar, ":"); idx != -1 {
			defaultVal = envVar[idx+1:]
			envVar = envVar[:idx]
		}

		val := os.Getenv(envVar)
		if val == "" {
			val = defaultVal
		}

		result = result[:start] + val + result[end+1:]
	}

	return result
}
