package zcfg

import (
	"fmt"
	"reflect"
	"sync"
)

// Config represents a configuration instance
type Config struct {
	rawMap  map[string]any
	target  any
	file    string
	option  *Option
	watcher *FileWatcher
	mu      sync.RWMutex
}

// Global registry to track configs by target type
var (
	registryMu     sync.RWMutex
	configRegistry = make(map[reflect.Type]*Config)
)

// MustLoad loads configuration from file, panics on error
func MustLoad[T any](file string, opts ...func(*Option)) *T {
	result, err := Load[T](file, opts...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustLoadFromJson loads configuration from JSON bytes, panics on error
func MustLoadFromJson[T any](content []byte, opts ...func(*Option)) *T {
	result, err := LoadFromJson[T](content, opts...)
	if err != nil {
		panic(err)
	}
	return result
}

// New creates a new Config instance
func New[T any](fn func(v *Config) error, opts ...func(*Option)) (*Config, error) {
	// Create target instance
	var target T
	v := &target

	// Create option
	option := NewOption()
	for _, opt := range opts {
		opt(option)
	}

	c := &Config{
		target: v,
		option: option,
	}

	// Call the function to setup c
	if err := fn(c); err != nil {
		return nil, err
	}
	if c.rawMap == nil {
		return nil, fmt.Errorf("rawMap is nil")
	}

	if option.UseEnv {
		loadEnvFile()
	}

	if err := mapToStruct(c.rawMap, v, option, false); err != nil {
		return nil, err
	}

	targetType := reflect.TypeOf(v)
	registryMu.Lock()
	configRegistry[targetType] = c
	registryMu.Unlock()

	// Setup hot reload if enabled
	if option.HotReload && option.Updatable && c.file != "" {
		watcher, err := NewFileWatcher(c)
		if err != nil {
			return nil, fmt.Errorf("failed to create file watcher: %w", err)
		}
		c.watcher = watcher
		if err := watcher.Start(); err != nil {
			return nil, fmt.Errorf("failed to start file watcher: %w", err)
		}
	}

	return c, nil
}

// Load loads configuration from file
func Load[T any](file string, opts ...func(*Option)) (*T, error) {
	config, err := New[T](func(v *Config) error {
		rawMap, err := parseConfigFile(file)
		if err != nil {
			return err
		}
		v.rawMap = rawMap
		v.file = file
		return nil
	}, opts...)

	if err != nil {
		return nil, err
	}

	return config.target.(*T), nil
}

// LoadFromJson loads configuration from JSON bytes
func LoadFromJson[T any](content []byte, opts ...func(*Option)) (*T, error) {
	config, err := New[T](func(v *Config) error {
		// Parse JSON content
		rawMap, err := parseJSONBytes(content)
		if err != nil {
			return err
		}
		v.rawMap = rawMap
		return nil
	}, opts...)

	if err != nil {
		return nil, err
	}

	return config.target.(*T), nil
}

// Get gets Config instance by target type
func Get[T any]() *Config {
	var target *T
	targetType := reflect.TypeOf(target)

	registryMu.RLock()
	config, exists := configRegistry[targetType]
	registryMu.RUnlock()

	if !exists {
		return nil
	}

	return config
}

// GetMap returns the raw configuration map
func (c *Config) GetMap() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]any)
	for k, v := range c.rawMap {
		result[k] = v
	}
	return result
}

// Update updates configuration with new map
func (c *Config) Update(m map[string]any) error {
	if !c.option.Updatable {
		return fmt.Errorf("config is not updatable")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update only the fields present in the update map
	if err := mapToStruct(m, c.target, c.option, true); err != nil {
		return fmt.Errorf("failed to update struct: %w", err)
	}

	return nil
}

// UpdateFromJson updates configuration from JSON bytes
func (c *Config) UpdateFromJson(content []byte) error {
	if !c.option.Updatable {
		return fmt.Errorf("config is not updatable")
	}

	// Parse JSON to map
	updateMap, err := parseJSONBytes(content)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Call Update method
	return c.Update(updateMap)
}

// GetValue gets value by path
func (c *Config) GetValue(path string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return getNestedValue(c.rawMap, path)
}

// StartWatcher starts the file watcher
func (c *Config) StartWatcher() error {
	if c.watcher == nil {
		return fmt.Errorf("no watcher configured")
	}

	if c.watcher.IsRunning() {
		return nil
	}

	return c.watcher.Start()
}

// StopWatcher stops the file watcher
func (c *Config) StopWatcher() error {
	if c.watcher == nil {
		return nil
	}

	return c.watcher.Stop()
}

// IsWatcherRunning returns whether the watcher is running
func (c *Config) IsWatcherRunning() bool {
	if c.watcher == nil {
		return false
	}

	return c.watcher.IsRunning()
}

// GetTarget returns the target struct pointer
func (c *Config) GetTarget() any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.target
}
