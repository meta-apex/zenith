package zcfg

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// Parse parses the configuration into the given struct
func (c *Config) Parse(v any, path string) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return c.parseStruct(val, c.data, path, false)
}

// parseStruct parses a struct value
func (c *Config) parseStruct(val reflect.Value, curMap map[string]any, path string, optional bool) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		typeField := typ.Field(i)
		tag := typeField.Tag.Get(c.options.TagName)
		if tag == "-" {
			continue
		}

		fieldName, opts := parseTag(tag)
		if fieldName == "" {
			fieldName = typeField.Name
		}

		// Handle embedded structs and struct fields
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			// Ensure we're working with a struct value
			structValue := field
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
				structValue = field.Elem()
			}

			if typeField.Anonymous {
				// For embedded structs, use the same map
				if err := c.parseStruct(structValue, curMap, path, false); err != nil {
					return err
				}
			} else {
				// For non-anonymous structs, get the nested map
				nestedMap, ok := c.getNestedMap(curMap, fieldName)
				if !ok {
					nestedMap = make(map[string]any)
				}
				if err := c.parseStruct(structValue, nestedMap, joinPath(path, fieldName), opts.optional); err != nil {
					return err
				}
			}
			continue
		}

		opts.path = path
		opts.parentOptional = optional
		if err := c.parseField(field, fieldName, opts, curMap); err != nil {
			return fmt.Errorf(`[%s%s]: %w`, path, fieldName, err)
		}
	}

	return nil
}

// parseField parses a single field
func (c *Config) parseField(field reflect.Value, key string, opts tagOptions, curMap map[string]any) error {
	var val any
	var ok bool

	// Try to get value from current map
	if curMap != nil {
		val, ok = c.getValueFromMap(curMap, key)
	}

	if !ok {
		if opts.optional {
			return nil
		}
		if !opts.hasDefault {
			if opts.parentOptional {
				return nil
			}
			return fmt.Errorf("required field is missing")
		}
		val = opts.defaultValue
	}

	// Handle environment variables
	if c.options.UseEnv {
		val = c.resolveEnvVars(val)
	}

	return c.setField(field, val, opts)
}

// getValueFromMap gets a value from the given map
func (c *Config) getValueFromMap(m map[string]any, key string) (any, bool) {
	if c.options.IgnoreCase {
		key = strings.ToLower(key)
	}

	if v, ok := m[key]; ok {
		return v, true
	}

	return nil, false
}

// getNestedMap gets a nested map from the configuration map
func (c *Config) getNestedMap(curMap map[string]any, key string) (map[string]any, bool) {
	if c.options.IgnoreCase {
		key = strings.ToLower(key)
	}

	if v, ok := curMap[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m, true
		}
	}

	return nil, false
}

// setField sets a field's value
func (c *Config) setField(field reflect.Value, val any, opts tagOptions) error {
	// Handle nil values
	if val == nil {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return fmt.Errorf("cannot set nil to non-pointer field")
	}

	// Handle special types
	switch field.Interface().(type) {
	case time.Duration:
		dur, err := cast.ToDurationE(val)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(dur))
		return nil
	}

	// Handle basic types
	switch field.Kind() {
	case reflect.String:
		str, err := cast.ToStringE(val)
		if err != nil {
			return err
		}
		if err := validateOptions(str, opts); err != nil {
			return err
		}
		field.SetString(str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := cast.ToInt64E(val)
		if err != nil {
			return err
		}
		if err := validateRange(n, opts); err != nil {
			return err
		}
		field.SetInt(n)

	case reflect.Float32, reflect.Float64:
		f, err := cast.ToFloat64E(val)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := cast.ToBoolE(val)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Slice:
		// 特殊处理字符串切片
		if field.Type().Elem().Kind() == reflect.String {
			switch v := val.(type) {
			case []interface{}:
				strSlice := make([]string, len(v))
				for i, item := range v {
					str, err := cast.ToStringE(item)
					if err != nil {
						return err
					}
					strSlice[i] = str
				}
				field.Set(reflect.ValueOf(strSlice))
			case []string:
				field.Set(reflect.ValueOf(v))
			default:
				return fmt.Errorf("cannot convert %T to []string", val)
			}

			return nil
		}

		// 处理其他类型的切片
		slice, err := cast.ToSliceE(val)
		if err != nil {
			return err
		}
		field.Set(reflect.MakeSlice(field.Type(), len(slice), len(slice)))
		for i, v := range slice {
			if err := c.setField(field.Index(i), v, tagOptions{}); err != nil {
				return err
			}
		}

	case reflect.Map:
		m, err := cast.ToStringMapE(val)
		if err != nil {
			return err
		}
		field.Set(reflect.MakeMap(field.Type()))
		for k, v := range m {
			keyValue := reflect.ValueOf(k)
			valValue := reflect.New(field.Type().Elem()).Elem()
			if err := c.setField(valValue, v, tagOptions{}); err != nil {
				return err
			}
			field.SetMapIndex(keyValue, valValue)
		}

	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return c.setField(field.Elem(), val, opts)

	case reflect.Struct:
		if m, ok := val.(map[string]any); ok {
			return c.parseStruct(field, m, opts.path, opts.optional)
		}
		return fmt.Errorf("cannot convert %T to struct", val)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

func joinPath(path, sep string) string {
	if path == "" {
		return sep + "."
	}
	return path + sep + "."
}
