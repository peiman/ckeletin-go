// internal/docs/generator.go

package docs

import (
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

// Generator handles document generation based on configuration
type Generator struct {
	cfg     Config
	appInfo AppInfo
}

// NewGenerator creates a new document generator with the given configuration
func NewGenerator(cfg Config) *Generator {
	return &Generator{
		cfg:     cfg,
		appInfo: AppInfo{}, // Empty AppInfo, will be populated when needed
	}
}

// SetAppInfo sets the application information used for documentation
func (g *Generator) SetAppInfo(info AppInfo) {
	g.appInfo = info
}

// Generate produces documentation in the configured format
func (g *Generator) Generate() error {
	var writer io.Writer = g.cfg.Writer
	var file io.WriteCloser
	var closeErr error

	// If output file is specified, create it
	if g.cfg.OutputFile != "" {
		var err error
		file, err = openOutputFile(g.cfg.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			// Capture the close error
			closeErr = file.Close()
			if closeErr != nil {
				log.Error().Err(closeErr).Str("file", g.cfg.OutputFile).Msg("Failed to close output file")
			}
		}()
		writer = file
		log.Info().Str("file", g.cfg.OutputFile).Msg("Writing documentation to file")
	}

	var err error
	switch strings.ToLower(g.cfg.OutputFormat) {
	case FormatMarkdown:
		err = g.GenerateMarkdownDocs(writer, g.appInfo)
	case FormatYAML:
		err = g.GenerateYAMLDocs(writer)
	default:
		err = fmt.Errorf("unsupported format: %s", g.cfg.OutputFormat)
	}

	// If there was no error from the operation but there was a close error, return the close error
	if err == nil && closeErr != nil {
		return fmt.Errorf("failed to close output file: %w", closeErr)
	}

	return err
}

// GenerateMarkdown is a convenience function to generate markdown documentation directly
func GenerateMarkdown(writer io.Writer, appInfo AppInfo) error {
	g := NewGenerator(NewConfig(writer, WithOutputFormat(FormatMarkdown)))
	g.SetAppInfo(appInfo)
	return g.GenerateMarkdownDocs(writer, appInfo)
}

// GenerateYAML is a convenience function to generate YAML configuration template directly
func GenerateYAML(writer io.Writer) error {
	g := NewGenerator(NewConfig(writer, WithOutputFormat(FormatYAML)))
	return g.GenerateYAMLDocs(writer)
}
