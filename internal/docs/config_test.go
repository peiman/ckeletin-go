// internal/docs/config_test.go

package docs

import (
	"bytes"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
)

func TestConfig(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}

	tests := []struct {
		name           string
		cfg            Config
		wantFormat     string
		wantOutputFile string
	}{
		{
			name: "Default config",
			cfg: Config{
				Writer:       writer,
				OutputFormat: FormatMarkdown,
				OutputFile:   "",
				Registry:     config.Registry,
			},
			wantFormat:     FormatMarkdown,
			wantOutputFile: "",
		},
		{
			name: "YAML format",
			cfg: Config{
				Writer:       writer,
				OutputFormat: FormatYAML,
				OutputFile:   "",
				Registry:     config.Registry,
			},
			wantFormat:     FormatYAML,
			wantOutputFile: "",
		},
		{
			name: "With output file",
			cfg: Config{
				Writer:       writer,
				OutputFormat: FormatMarkdown,
				OutputFile:   "test.md",
				Registry:     config.Registry,
			},
			wantFormat:     FormatMarkdown,
			wantOutputFile: "test.md",
		},
		{
			name: "YAML with output file",
			cfg: Config{
				Writer:       writer,
				OutputFormat: FormatYAML,
				OutputFile:   "test.yaml",
				Registry:     config.Registry,
			},
			wantFormat:     FormatYAML,
			wantOutputFile: "test.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// ASSERTION PHASE
			if tt.cfg.OutputFormat != tt.wantFormat {
				t.Errorf("Expected format %s, got %s", tt.wantFormat, tt.cfg.OutputFormat)
			}
			if tt.cfg.OutputFile != tt.wantOutputFile {
				t.Errorf("Expected output file %s, got %s", tt.wantOutputFile, tt.cfg.OutputFile)
			}
			if tt.cfg.Writer != writer {
				t.Errorf("Expected writer to be set correctly")
			}
		})
	}
}
