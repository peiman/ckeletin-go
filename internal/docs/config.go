// internal/docs/config.go

package docs

import (
	"io"
	"os"

	"github.com/peiman/ckeletin-go/internal/config"
)

// Define output format types
const (
	FormatMarkdown = "markdown"
	FormatYAML     = "yaml"
)

// RegistryFunc defines a function that returns a configuration registry
// This is primarily used for testing to mock the registry
type RegistryFunc func() []config.ConfigOption

// Config holds all configuration for document generation
// This struct uses the Options Pattern for testability and clarity
type Config struct {
	OutputFormat string
	OutputFile   string
	Writer       io.Writer
	Registry     RegistryFunc
}

// Option is a function that modifies Config
type Option func(*Config)

// WithOutputFormat sets the output format for document generation
func WithOutputFormat(format string) Option {
	return func(cfg *Config) { cfg.OutputFormat = format }
}

// WithOutputFile sets the output file path for document generation
func WithOutputFile(file string) Option {
	return func(cfg *Config) { cfg.OutputFile = file }
}

// WithWriter sets a custom writer for document generation
func WithWriter(writer io.Writer) Option {
	return func(cfg *Config) { cfg.Writer = writer }
}

// WithRegistryFunc sets a custom registry function for testing
func WithRegistryFunc(fn RegistryFunc) Option {
	return func(cfg *Config) { cfg.Registry = fn }
}

// NewConfig creates a new Config with default values and applies options
func NewConfig(defaultWriter io.Writer, opts ...Option) Config {
	// Default configuration
	cfg := Config{
		OutputFormat: FormatMarkdown,
		OutputFile:   "",
		Writer:       defaultWriter,
		Registry:     config.Registry,
	}

	// Apply all options
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// Variable to facilitate testing file operations
var openOutputFile = func(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}
