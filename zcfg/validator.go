package zcfg

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TagInfo represents parsed tag information
type TagInfo struct {
	FieldName string   // Custom field name
	Default   string   // Default value
	Options   []string // Valid options
	RangeMin  *float64 // Range minimum
	RangeMax  *float64 // Range maximum
	Watch     bool     // Whether to watch for changes
	Optional  bool     // Whether field is optional
	Skip      bool     // Whether to skip this field
}

// parseTag parses struct tag and returns TagInfo
func parseTag(tag string) *TagInfo {
	info := &TagInfo{}

	if tag == "" {
		return info
	}

	// Handle skip tag
	if tag == "_" {
		info.Skip = true
		return info
	}

	parts := strings.Split(tag, ",")

	// First part might be field name
	if len(parts) > 0 && parts[0] != "" {
		info.FieldName = parts[0]
	}

	// Parse other parts
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}

		switch {
		case strings.HasPrefix(part, "default="):
			info.Default = strings.TrimPrefix(part, "default=")
		case strings.HasPrefix(part, "options="):
			optionsStr := strings.TrimPrefix(part, "options=")
			info.Options = strings.Split(optionsStr, "|")
		case strings.HasPrefix(part, "range="):
			parseRange(strings.TrimPrefix(part, "range="), info)
		case part == "watch":
			info.Watch = true
		case part == "optional":
			info.Optional = true
		}
	}

	return info
}

// parseRange parses range specification like [0:100], (0:100], [0:100), (0:100)
func parseRange(rangeStr string, info *TagInfo) {
	rangeRegex := regexp.MustCompile(`^([\[\(])([^:]*):([^\]\)]*)([\]\)])$`)
	matches := rangeRegex.FindStringSubmatch(rangeStr)
	if len(matches) != 5 {
		return
	}

	minStr := strings.TrimSpace(matches[2])
	maxStr := strings.TrimSpace(matches[3])

	// Parse min value
	if minStr != "" {
		if fMin, err := strconv.ParseFloat(minStr, 64); err == nil {
			info.RangeMin = &fMin
		}
	}

	// Parse max value
	if maxStr != "" {
		if fMax, err := strconv.ParseFloat(maxStr, 64); err == nil {
			info.RangeMax = &fMax
		}
	}
}

// validateValue validates value according to tag rules
func validateValue(value any, tagInfo *TagInfo, fieldPath string) error {
	if tagInfo.Skip || tagInfo.Optional {
		return nil
	}

	// Validate options
	if len(tagInfo.Options) > 0 {
		valueStr := fmt.Sprintf("%v", value)
		valid := false
		for _, option := range tagInfo.Options {
			if valueStr == option {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("field %s value '%v' is not in valid options: %v", fieldPath, value, tagInfo.Options)
		}
	}

	// Validate range for numeric values
	if tagInfo.RangeMin != nil || tagInfo.RangeMax != nil {
		if err := validateRange(value, tagInfo, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

// validateRange validates numeric range
func validateRange(value any, tagInfo *TagInfo, fieldPath string) error {
	var numValue float64
	var err error

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int8:
		numValue = float64(v)
	case int16:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case uint:
		numValue = float64(v)
	case uint8:
		numValue = float64(v)
	case uint16:
		numValue = float64(v)
	case uint32:
		numValue = float64(v)
	case uint64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	case string:
		numValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("field %s value '%v' is not a valid number for range validation", fieldPath, value)
		}
	default:
		return fmt.Errorf("field %s value '%v' is not a numeric type for range validation", fieldPath, value)
	}

	// Check minimum
	if tagInfo.RangeMin != nil {
		min := *tagInfo.RangeMin
		if numValue < min {
			return fmt.Errorf("field %s value %v must be greater than or equal to %v", fieldPath, numValue, min)
		}
	}

	// Check maximum
	if tagInfo.RangeMax != nil {
		max := *tagInfo.RangeMax
		if numValue > max {
			return fmt.Errorf("field %s value %v must be less than or equal to %v", fieldPath, numValue, max)
		}
	}

	return nil
}

// parseDuration parses duration string, supporting both time.Duration format and milliseconds
func parseDuration(value any) (time.Duration, error) {
	switch v := value.(type) {
	case string:
		// Try parsing as time.Duration first
		if duration, err := time.ParseDuration(v); err == nil {
			return duration, nil
		}
		// Try parsing as number (milliseconds)
		if ms, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(ms) * time.Millisecond, nil
		}
		return 0, fmt.Errorf("invalid duration format: %s", v)
	case int, int8, int16, int32, int64:
		// Treat as milliseconds
		ms := reflect.ValueOf(v).Int()
		return time.Duration(ms) * time.Millisecond, nil
	case uint, uint8, uint16, uint32, uint64:
		// Treat as milliseconds
		ms := reflect.ValueOf(v).Uint()
		return time.Duration(ms) * time.Millisecond, nil
	case float32, float64:
		// Treat as milliseconds
		ms := reflect.ValueOf(v).Float()
		return time.Duration(ms) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("unsupported duration type: %T", v)
	}
}
