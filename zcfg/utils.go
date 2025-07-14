package zcfg

import (
	"regexp"
	"strings"
	"unicode"
)

// convertFieldName converts field name according to match mode
func convertFieldName(fieldName string, mode MatchMode) string {
	switch mode {
	case MatchNormal:
		return fieldName
	case MatchIgnoreCase:
		return strings.ToLower(fieldName)
	case MatchCamelCase:
		return toCamelCase(fieldName)
	case MatchSnakeCase:
		return toSnakeCase(fieldName)
	default:
		return strings.ToLower(fieldName)
	}
}

// toCamelCase converts PascalCase to camelCase
func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Insert underscore before uppercase letters (except the first one)
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(s, `${1}_${2}`)

	// Convert to lowercase
	return strings.ToLower(snake)
}

// findValueInMap finds value in map using different matching modes
func findValueInMap(m map[string]any, fieldName string, mode MatchMode) (any, bool) {
	// First try exact match
	if value, exists := m[fieldName]; exists {
		return value, true
	}

	// Then try converted field name
	convertedName := convertFieldName(fieldName, mode)
	if value, exists := m[convertedName]; exists {
		return value, true
	}

	// For case insensitive mode, try all keys
	if mode == MatchIgnoreCase {
		lowerFieldName := strings.ToLower(fieldName)
		for key, value := range m {
			if strings.ToLower(key) == lowerFieldName {
				return value, true
			}
		}
	}

	return nil, false
}

// getNestedValue gets nested value from map using dot notation path
func getNestedValue(m map[string]any, path string) (any, bool) {
	if path == "" {
		return m, true
	}

	parts := strings.Split(path, ".")
	current := any(m)

	for _, part := range parts {
		if currentMap, ok := current.(map[string]any); ok {
			if value, exists := currentMap[part]; exists {
				current = value
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

// setNestedValue sets nested value in map using dot notation path
func setNestedValue(m map[string]any, path string, value any) {
	if path == "" {
		return
	}

	parts := strings.Split(path, ".")
	current := m

	for _, part := range parts[:len(parts)-1] {
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}

		if nextMap, ok := current[part].(map[string]any); ok {
			current = nextMap
		} else {
			// Create new nested map
			newMap := make(map[string]any)
			current[part] = newMap
			current = newMap
		}
	}

	lastPart := parts[len(parts)-1]
	current[lastPart] = value
}

// isZeroValue checks if a value is zero value
func isZeroValue(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return val == ""
	case int, int8, int16, int32, int64:
		return val == 0
	case uint, uint8, uint16, uint32, uint64:
		return val == 0
	case float32, float64:
		return val == 0
	case bool:
		return !val
	default:
		return false
	}
}
