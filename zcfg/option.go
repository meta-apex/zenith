package zcfg

// Options represents global configuration options
type Options struct {
	// IgnoreCase determines whether to ignore case when matching field names
	IgnoreCase bool
	// TagName specifies the struct tag name to use (default: "meta")
	TagName string
	// UseDefaults determines whether to use default values from struct tags
	UseDefaults bool
	// UseEnv determines whether to use environment variables
	UseEnv bool
}

// Option represents a function that can be used to configure a Config instance
type Option func(*Config)

// WithIgnoreCase sets whether to ignore case when matching field names
func WithIgnoreCase(ignore bool) Option {
	return func(c *Config) {
		c.options.IgnoreCase = ignore
	}
}

// WithTagName sets the struct tag name to use
func WithTagName(name string) Option {
	return func(c *Config) {
		c.options.TagName = name
	}
}

// WithDefaults sets whether to use default values from struct tags
func WithDefaults(use bool) Option {
	return func(c *Config) {
		c.options.UseDefaults = use
	}
}

// WithEnv sets whether to use environment variables
func WithEnv(use bool) Option {
	return func(c *Config) {
		c.options.UseEnv = use
	}
}
