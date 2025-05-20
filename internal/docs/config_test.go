// internal/docs/config_test.go

package docs

import (
	"bytes"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}

	// EXECUTION PHASE
	cfg := NewConfig(writer)

	// ASSERTION PHASE
	if cfg.OutputFormat != FormatMarkdown {
		t.Errorf("Expected default format to be %s, got %s", FormatMarkdown, cfg.OutputFormat)
	}
	if cfg.OutputFile != "" {
		t.Errorf("Expected default output file to be empty, got %s", cfg.OutputFile)
	}
	if cfg.Writer != writer {
		t.Errorf("Expected writer to be set correctly")
	}
}

func TestConfigOptions(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}

	tests := []struct {
		name           string
		options        []Option
		wantFormat     string
		wantOutputFile string
	}{
		{
			name:           "WithOutputFormat option",
			options:        []Option{WithOutputFormat(FormatYAML)},
			wantFormat:     FormatYAML,
			wantOutputFile: "",
		},
		{
			name:           "WithOutputFile option",
			options:        []Option{WithOutputFile("test.md")},
			wantFormat:     FormatMarkdown, // Default
			wantOutputFile: "test.md",
		},
		{
			name:           "Multiple options",
			options:        []Option{WithOutputFormat(FormatYAML), WithOutputFile("test.yaml")},
			wantFormat:     FormatYAML,
			wantOutputFile: "test.yaml",
		},
		{
			name:           "WithWriter option",
			options:        []Option{WithWriter(new(bytes.Buffer))},
			wantFormat:     FormatMarkdown, // Default
			wantOutputFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// EXECUTION PHASE
			cfg := NewConfig(writer, tt.options...)

			// ASSERTION PHASE
			if cfg.OutputFormat != tt.wantFormat {
				t.Errorf("Expected format %s, got %s", tt.wantFormat, cfg.OutputFormat)
			}
			if cfg.OutputFile != tt.wantOutputFile {
				t.Errorf("Expected output file %s, got %s", tt.wantOutputFile, cfg.OutputFile)
			}
		})
	}
}
