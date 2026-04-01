package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{name: "valid simple name", input: "greet", wantErr: false},
		{name: "valid with digits", input: "cmd2", wantErr: false},
		{name: "valid with underscore", input: "my_cmd", wantErr: false},
		{name: "empty name", input: "", wantErr: true, errMsg: "cannot be empty"},
		{name: "starts with digit", input: "2cmd", wantErr: true, errMsg: "cannot start with a digit"},
		{name: "contains hyphen", input: "my-cmd", wantErr: true, errMsg: "invalid character"},
		{name: "contains space", input: "my cmd", wantErr: true, errMsg: "invalid character"},
		{name: "reserved root", input: "root", wantErr: true, errMsg: "reserved"},
		{name: "reserved helpers", input: "helpers", wantErr: true, errMsg: "reserved"},
		{name: "reserved flags", input: "flags", wantErr: true, errMsg: "reserved"},
		{name: "reserved version", input: "version", wantErr: true, errMsg: "reserved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "simple", input: "greet", expect: "Greet"},
		{name: "underscore", input: "my_cmd", expect: "MyCmd"},
		{name: "single char", input: "a", expect: "A"},
		{name: "already capitalized segments", input: "Hello_World", expect: "HelloWorld"},
		{name: "multiple underscores", input: "a_b_c", expect: "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, toTitle(tt.input))
		})
	}
}

func TestReadModulePath(t *testing.T) {
	t.Run("valid go.mod", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		goModPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(goModPath, []byte("module github.com/example/myapp\n\ngo 1.21\n"), 0o644)
		require.NoError(t, err)

		mod, err := readModulePath(goModPath)
		require.NoError(t, err)
		assert.Equal(t, "github.com/example/myapp", mod)
	})

	t.Run("missing go.mod", func(t *testing.T) {
		t.Parallel()
		_, err := readModulePath("/nonexistent/go.mod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot open")
	})

	t.Run("go.mod without module directive", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		goModPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(goModPath, []byte("go 1.21\n"), 0o644)
		require.NoError(t, err)

		_, err = readModulePath(goModPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no module directive")
	})
}

func TestFilesToGenerate(t *testing.T) {
	t.Parallel()
	data := templateData{
		Name:       "greet",
		NameTitle:  "Greet",
		ModulePath: "github.com/example/myapp",
	}

	files := filesToGenerate(data)
	require.Len(t, files, 4)

	assert.Equal(t, filepath.Join("cmd", "greet.go"), files[0].path)
	assert.Equal(t, filepath.Join("internal", "greet", "greet.go"), files[1].path)
	assert.Equal(t, filepath.Join("internal", "greet", "greet_test.go"), files[2].path)
	assert.Equal(t, filepath.Join("internal", "config", "commands", "greet_config.go"), files[3].path)
}

func TestGenerateFile_CreatesValidGoFiles(t *testing.T) {
	data := templateData{
		Name:       "greet",
		NameTitle:  "Greet",
		ModulePath: "github.com/example/myapp",
	}

	files := filesToGenerate(data)
	dir := t.TempDir()

	fset := token.NewFileSet()

	for _, f := range files {
		t.Run(f.tmplName, func(t *testing.T) {
			outPath := filepath.Join(dir, f.path)
			err := generateFile(outPath, f.tmplName, f.tmpl, data)
			require.NoError(t, err, "generateFile should succeed for %s", f.tmplName)

			// Verify the file exists
			info, err := os.Stat(outPath)
			require.NoError(t, err, "generated file should exist")
			assert.True(t, info.Size() > 0, "generated file should not be empty")

			// Verify the file parses as valid Go
			_, err = parser.ParseFile(fset, outPath, nil, parser.AllErrors)
			assert.NoError(t, err, "generated file %s should be valid Go", f.path)
		})
	}
}

func TestGenerateFile_FailsIfFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.go")

	// Create the file first
	err := os.WriteFile(path, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	data := templateData{
		Name:       "greet",
		NameTitle:  "Greet",
		ModulePath: "github.com/example/myapp",
	}

	err = generateFile(path, "cmd", cmdTemplate, data)
	require.Error(t, err, "should fail when file already exists")
	assert.Contains(t, err.Error(), "exists.go")
}

func TestGenerateFile_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "file.go")

	data := templateData{
		Name:       "greet",
		NameTitle:  "Greet",
		ModulePath: "github.com/example/myapp",
	}

	err := generateFile(path, "executor", executorTemplate, data)
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err, "file should exist in nested directory")
}

func TestGeneratedContent_ContainsExpectedPatterns(t *testing.T) {
	data := templateData{
		Name:       "greet",
		NameTitle:  "Greet",
		ModulePath: "github.com/example/myapp",
	}

	files := filesToGenerate(data)
	dir := t.TempDir()

	for _, f := range files {
		outPath := filepath.Join(dir, f.path)
		err := generateFile(outPath, f.tmplName, f.tmpl, data)
		require.NoError(t, err)
	}

	t.Run("cmd file uses MustNewCommand", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(dir, "cmd", "greet.go"))
		require.NoError(t, err)
		s := string(content)
		assert.Contains(t, s, "MustNewCommand(commands.GreetMetadata")
		assert.Contains(t, s, "MustAddToRoot(greetCmd)")
		assert.Contains(t, s, "greet.NewExecutor")
		assert.Contains(t, s, "greet.Config{")
	})

	t.Run("executor has Config and Executor types", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(dir, "internal", "greet", "greet.go"))
		require.NoError(t, err)
		s := string(content)
		assert.Contains(t, s, "type Config struct")
		assert.Contains(t, s, "type Executor struct")
		assert.Contains(t, s, "func NewExecutor(")
		assert.Contains(t, s, "func (e *Executor) Execute() error")
		assert.Contains(t, s, `"github.com/rs/zerolog/log"`)
	})

	t.Run("test file uses testify", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(dir, "internal", "greet", "greet_test.go"))
		require.NoError(t, err)
		s := string(content)
		assert.Contains(t, s, `"github.com/stretchr/testify/assert"`)
		assert.Contains(t, s, `"github.com/stretchr/testify/require"`)
		assert.Contains(t, s, "func TestNewExecutor(")
		assert.Contains(t, s, "func TestExecutor_Execute(")
		assert.Contains(t, s, "t.Parallel()")
	})

	t.Run("config file has metadata and provider", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(dir, "internal", "config", "commands", "greet_config.go"))
		require.NoError(t, err)
		s := string(content)
		assert.Contains(t, s, "var GreetMetadata = config.CommandMetadata{")
		assert.Contains(t, s, "func GreetOptions() []config.ConfigOption")
		assert.Contains(t, s, "config.RegisterOptionsProvider(GreetOptions)")
		assert.Contains(t, s, `ConfigPrefix: "app.greet"`)
	})
}
