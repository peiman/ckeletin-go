// Command generator for ckeletin-go
//
// Generates the full command pattern:
//   - cmd/<name>.go             — ultra-thin wrapper using MustNewCommand/MustAddToRoot
//   - internal/<name>/<name>.go — Executor with Config struct, NewExecutor, Execute
//   - internal/<name>/<name>_test.go — table-driven tests with testify
//   - internal/config/commands/<name>_config.go — CommandMetadata + RegisterOptionsProvider
//
// Usage:
//
//	go run ./.ckeletin/scripts/generate-command/ <name>
//
// Example:
//
//	go run ./.ckeletin/scripts/generate-command/ greet
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command-name>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s greet\n", os.Args[0])
		os.Exit(1)
	}

	name := strings.TrimSpace(os.Args[1])
	if err := validateName(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	modulePath, err := readModulePath("go.mod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading module path: %v\n", err)
		os.Exit(1)
	}

	data := templateData{
		Name:       name,
		NameTitle:  toTitle(name),
		ModulePath: modulePath,
	}

	files := filesToGenerate(data)

	// Check that none of the target files already exist
	for _, f := range files {
		if _, err := os.Stat(f.path); err == nil {
			fmt.Fprintf(os.Stderr, "Error: file already exists: %s\n", f.path)
			os.Exit(1)
		}
	}

	// Generate all files
	for _, f := range files {
		if err := generateFile(f.path, f.tmplName, f.tmpl, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", f.path, err)
			os.Exit(1)
		}
		fmt.Printf("  created %s\n", f.path)
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Run: task generate:config:key-constants\n")
	fmt.Printf("  2. Run: task format\n")
	fmt.Printf("  3. Customize business logic in internal/%s/%s.go\n", name, name)
	fmt.Printf("  4. Update metadata/options in internal/config/commands/%s_config.go\n", name)
	fmt.Printf("  5. Run: task check\n")
}

// templateData holds the data passed to all templates.
type templateData struct {
	Name       string // lowercase command name, e.g. "greet"
	NameTitle  string // title-cased command name, e.g. "Greet"
	ModulePath string // Go module path, e.g. "github.com/peiman/ckeletin-go"
}

// fileSpec describes a file to generate.
type fileSpec struct {
	path     string
	tmplName string
	tmpl     string
}

// filesToGenerate returns the list of files that should be created.
func filesToGenerate(data templateData) []fileSpec {
	return []fileSpec{
		{
			path:     filepath.Join("cmd", data.Name+".go"),
			tmplName: "cmd",
			tmpl:     cmdTemplate,
		},
		{
			path:     filepath.Join("internal", data.Name, data.Name+".go"),
			tmplName: "executor",
			tmpl:     executorTemplate,
		},
		{
			path:     filepath.Join("internal", data.Name, data.Name+"_test.go"),
			tmplName: "executor_test",
			tmpl:     executorTestTemplate,
		},
		{
			path:     filepath.Join("internal", "config", "commands", data.Name+"_config.go"),
			tmplName: "config",
			tmpl:     configTemplate,
		},
	}
}

// validateName checks that the command name is valid.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("command name %q contains invalid character %q; use only letters, digits, and underscores", name, r)
		}
	}
	if unicode.IsDigit(rune(name[0])) {
		return fmt.Errorf("command name %q cannot start with a digit", name)
	}
	// Disallow names that would conflict with existing framework files
	reserved := map[string]bool{
		"root": true, "helpers": true, "flags": true, "version": true,
	}
	if reserved[name] {
		return fmt.Errorf("command name %q is reserved", name)
	}
	return nil
}

// readModulePath extracts the module path from go.mod.
func readModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", goModPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading %s: %w", goModPath, err)
	}
	return "", fmt.Errorf("no module directive found in %s", goModPath)
}

// toTitle converts a lowercase name to TitleCase. Handles underscores by
// capitalizing each segment: "my_cmd" -> "MyCmd".
func toTitle(name string) string {
	parts := strings.Split(name, "_")
	var b strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		b.WriteRune(unicode.ToUpper(rune(part[0])))
		b.WriteString(part[1:])
	}
	return b.String()
}

// generateFile renders a template and writes it to disk, creating parent
// directories as needed.
func generateFile(path, tmplName, tmplText string, data templateData) error {
	t, err := template.New(tmplName).Parse(tmplText)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplName, err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", tmplName, err)
	}

	return nil
}

// ─── Templates ──────────────────────────────────────────────────────

// cmdTemplate generates the ultra-thin command wrapper in cmd/<name>.go
var cmdTemplate = `// cmd/{{.Name}}.go

package cmd

import (
	"{{.ModulePath}}/internal/config/commands"
	"{{.ModulePath}}/internal/{{.Name}}"
	"github.com/spf13/cobra"
)

var {{.Name}}Cmd = MustNewCommand(commands.{{.NameTitle}}Metadata, run{{.NameTitle}})

func init() {
	MustAddToRoot({{.Name}}Cmd)
}

func run{{.NameTitle}}(cmd *cobra.Command, args []string) error {
	cfg := {{.Name}}.Config{
		// TODO: Map config values to Config struct fields.
		// Example:
		//   Message: getConfigValueWithFlags[string](cmd, "message", config.KeyApp{{.NameTitle}}Message),
	}
	return {{.Name}}.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
`

// executorTemplate generates the Executor pattern in internal/<name>/<name>.go
var executorTemplate = `// internal/{{.Name}}/{{.Name}}.go

package {{.Name}}

import (
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
)

// Config holds configuration for the {{.Name}} command.
type Config struct {
	// TODO: Add configuration fields.
	// Example:
	//   Message string
}

// Executor handles the execution of the {{.Name}} command.
type Executor struct {
	cfg    Config
	writer io.Writer
}

// NewExecutor creates a new {{.Name}} command executor.
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	return &Executor{
		cfg:    cfg,
		writer: writer,
	}
}

// Execute runs the {{.Name}} command logic.
func (e *Executor) Execute() error {
	log.Debug().Str("component", "{{.Name}}").Msg("Starting {{.Name}} execution")

	// TODO: Implement command logic.
	fmt.Fprintln(e.writer, "{{.Name}}: not yet implemented")

	return nil
}
`

// executorTestTemplate generates table-driven tests in internal/<name>/<name>_test.go
var executorTestTemplate = `// internal/{{.Name}}/{{.Name}}_test.go

package {{.Name}}

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	t.Parallel()
	cfg := Config{}
	buf := &bytes.Buffer{}
	executor := NewExecutor(cfg, buf)
	assert.NotNil(t, executor)
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "Default execution",
			cfg:        Config{},
			wantOutput: "{{.Name}}: not yet implemented\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			executor := NewExecutor(tt.cfg, buf)

			err := executor.Execute()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}
`

// configTemplate generates the config metadata in internal/config/commands/<name>_config.go
var configTemplate = `// internal/config/commands/{{.Name}}_config.go
//
// {{.NameTitle}} command configuration: metadata + options
//
// After defining options here, run ` + "`task generate:config:key-constants`" + ` to
// generate type-safe constants. Then use them in your business logic.

package commands

import "{{.ModulePath}}/.ckeletin/pkg/config"

// {{.NameTitle}}Metadata defines all metadata for the {{.Name}} command.
var {{.NameTitle}}Metadata = config.CommandMetadata{
	Use:          "{{.Name}}",
	Short:        "TODO: Short description for {{.Name}}",
	Long:         "TODO: Long description for {{.Name}}.",
	ConfigPrefix: "app.{{.Name}}",
	FlagOverrides: map[string]string{
		// TODO: Map config keys to flag names.
		// Example:
		//   "app.{{.Name}}.message": "message",
	},
	Examples: []string{
		"{{.Name}}",
	},
}

// {{.NameTitle}}Options returns configuration options for the {{.Name}} command.
func {{.NameTitle}}Options() []config.ConfigOption {
	return []config.ConfigOption{
		// TODO: Add configuration options.
		// Example:
		// {
		//     Key:          "app.{{.Name}}.message",
		//     DefaultValue: "Hello",
		//     Description:  "Message to display",
		//     Type:         "string",
		//     ShortFlag:    "m",
		//     Required:     false,
		//     Example:      "Hello World!",
		// },
	}
}

// Self-register {{.Name}} options provider at init time.
func init() {
	config.RegisterOptionsProvider({{.NameTitle}}Options)
}
`
