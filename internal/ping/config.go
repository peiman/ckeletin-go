// internal/ping/config.go

package ping

// Config holds all configuration for the ping command
// This struct uses the Options Pattern for testability and clarity
type Config struct {
	Message string
	Color   string
	UI      bool
}

// Option is a function that modifies Config
type Option func(*Config)

// WithMessage sets the message to display
func WithMessage(msg string) Option {
	return func(cfg *Config) { cfg.Message = msg }
}

// WithColor sets the color for the output
func WithColor(color string) Option {
	return func(cfg *Config) { cfg.Color = color }
}

// WithUI sets whether to use the interactive UI
func WithUI(ui bool) Option {
	return func(cfg *Config) { cfg.UI = ui }
}

// NewConfig creates a new Config with default values and applies options
func NewConfig(message, color string, ui bool, opts ...Option) Config {
	// Default configuration from parameters
	cfg := Config{
		Message: message,
		Color:   color,
		UI:      ui,
	}

	// Apply all options
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
