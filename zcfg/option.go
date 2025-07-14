package zcfg

// MatchMode represents field name matching mode
type MatchMode int

const (
	MatchNormal     MatchMode = iota // Normal matching
	MatchIgnoreCase                  // Case insensitive matching
	MatchCamelCase                   // Camel case matching
	MatchSnakeCase                   // Snake case matching
)

// WatchCallback is the callback function for field changes
type WatchCallback func(path, key string, oldValue, newValue any) error

// Option represents configuration options
type Option struct {
	TagName       string        // Tag name, default "meta"
	MatchMode     MatchMode     // Field matching mode
	UseEnv        bool          // Whether to use environment variables
	Updatable     bool          // Whether to support updates
	HotReload     bool          // Whether to enable hot reload
	WatchCallback WatchCallback // Watch callback function
}

// NewOption creates a new Option with default values
func NewOption() *Option {
	return &Option{
		TagName:       "meta",
		MatchMode:     MatchIgnoreCase,
		UseEnv:        true,
		Updatable:     false,
		HotReload:     false,
		WatchCallback: nil,
	}
}

// WithTagName sets the tag name
func WithTagName(tagName string) func(*Option) {
	return func(o *Option) {
		o.TagName = tagName
	}
}

// WithMatchMode sets the field matching mode
func WithMatchMode(mode MatchMode) func(*Option) {
	return func(o *Option) {
		o.MatchMode = mode
	}
}

// WithUseEnv sets whether to use environment variables
func WithUseEnv(useEnv bool) func(*Option) {
	return func(o *Option) {
		o.UseEnv = useEnv
	}
}

// WithUpdatable sets whether to support updates
func WithUpdatable(updatable bool) func(*Option) {
	return func(o *Option) {
		o.Updatable = updatable
	}
}

// WithHotReload sets whether to enable hot reload
func WithHotReload(hotReload bool) func(*Option) {
	return func(o *Option) {
		o.HotReload = hotReload
	}
}

// WithWatchCallback sets the watch callback function
func WithWatchCallback(callback WatchCallback) func(*Option) {
	return func(o *Option) {
		o.WatchCallback = callback
	}
}
