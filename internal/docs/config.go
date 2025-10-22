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
type Config struct {
	OutputFormat string
	OutputFile   string
	Writer       io.Writer
	Registry     RegistryFunc
}

// Variable to facilitate testing file operations
var openOutputFile = func(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}
