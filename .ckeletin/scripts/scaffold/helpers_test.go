package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceModulePreservingPkg(t *testing.T) {
	const oldModule = "github.com/peiman/ckeletin-go"
	const newModule = "github.com/user/myapp"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces standard internal import",
			input:    `import "github.com/peiman/ckeletin-go/internal/check"`,
			expected: `import "github.com/user/myapp/internal/check"`,
		},
		{
			name:     "replaces .ckeletin import",
			input:    `import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"`,
			expected: `import "github.com/user/myapp/.ckeletin/pkg/config"`,
		},
		{
			name:     "preserves pkg/checkmate import",
			input:    `	"github.com/peiman/ckeletin-go/pkg/checkmate"`,
			expected: `	"github.com/peiman/ckeletin-go/pkg/checkmate"`,
		},
		{
			name:     "preserves any pkg/ import",
			input:    `	"github.com/peiman/ckeletin-go/pkg/somefuture"`,
			expected: `	"github.com/peiman/ckeletin-go/pkg/somefuture"`,
		},
		{
			name: "handles mixed import block",
			input: `import (
	"fmt"
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
)`,
			expected: `import (
	"fmt"
	"github.com/user/myapp/internal/ping"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/user/myapp/.ckeletin/pkg/config"
)`,
		},
		{
			name:     "replaces module in go.mod line",
			input:    `module github.com/peiman/ckeletin-go`,
			expected: `module github.com/user/myapp`,
		},
		{
			name:     "replaces module reference in comments",
			input:    `// See github.com/peiman/ckeletin-go/internal/check for details`,
			expected: `// See github.com/user/myapp/internal/check for details`,
		},
		{
			name:     "preserves pkg/ reference in comments",
			input:    `// Uses github.com/peiman/ckeletin-go/pkg/checkmate for output`,
			expected: `// Uses github.com/peiman/ckeletin-go/pkg/checkmate for output`,
		},
		{
			name:     "no changes when no module reference",
			input:    `package main`,
			expected: `package main`,
		},
		{
			name:     "handles empty input",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceModulePreservingPkg(tt.input, oldModule, newModule)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestToEnvPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"hyphenated name", "ckeletin-go", "CKELETIN_GO"},
		{"simple name", "myapp", "MYAPP"},
		{"dotted name", "app.v2", "APP_V2"},
		{"already uppercase", "MYAPP", "MYAPP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, toEnvPrefix(tt.input))
		})
	}
}

func TestParseModuleParts(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "standard github module",
			module:    "github.com/me/myproject",
			wantOwner: "me",
			wantRepo:  "myproject",
		},
		{
			name:      "gitlab module",
			module:    "gitlab.com/team/service",
			wantOwner: "team",
			wantRepo:  "service",
		},
		{
			name:      "vanity domain module",
			module:    "example.com/org/tool",
			wantOwner: "org",
			wantRepo:  "tool",
		},
		{
			name:      "deep path module uses first two segments after host",
			module:    "github.com/org/repo/v2",
			wantOwner: "org",
			wantRepo:  "repo",
		},
		{
			name:      "two-segment module (no host prefix in standard sense)",
			module:    "example.com/tool",
			wantOwner: "",
			wantRepo:  "tool",
		},
		{
			name:      "single segment module",
			module:    "mymodule",
			wantOwner: "",
			wantRepo:  "mymodule",
		},
		{
			name:      "empty string module",
			module:    "",
			wantOwner: "",
			wantRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo := parseModuleParts(tt.module)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}

func TestResetChangelog(t *testing.T) {
	t.Run("creates clean keepachangelog template", func(t *testing.T) {
		tmpDir := t.TempDir()
		changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

		// Write existing changelog with ckeletin-go history
		require.NoError(t, os.WriteFile(changelogPath, []byte("# Changelog\n\n## [0.9.1] - old stuff\n"), 0600))

		err := resetChangelog(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(changelogPath)
		require.NoError(t, err)

		got := string(content)
		assert.Contains(t, got, "# Changelog")
		assert.Contains(t, got, "Keep a Changelog")
		assert.Contains(t, got, "Semantic Versioning")
		assert.Contains(t, got, "## [Unreleased]")
		assert.NotContains(t, got, "0.9.1")
		assert.NotContains(t, got, "old stuff")
	})

	t.Run("creates changelog even if none exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := resetChangelog(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "CHANGELOG.md"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "## [Unreleased]")
	})
}

func TestResetLicense(t *testing.T) {
	t.Run("replaces existing license with MIT template", func(t *testing.T) {
		tmpDir := t.TempDir()
		licensePath := filepath.Join(tmpDir, "LICENSE")

		require.NoError(t, os.WriteFile(licensePath, []byte("Copyright Peiman Khorramshahi"), 0600))

		err := resetLicense(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(licensePath)
		require.NoError(t, err)

		got := string(content)
		assert.Contains(t, got, "MIT License")
		assert.Contains(t, got, "[YEAR]")
		assert.Contains(t, got, "[YOUR NAME OR COMPANY]")
		assert.Contains(t, got, "Permission is hereby granted")
		assert.NotContains(t, got, "Peiman")
	})

	t.Run("creates license even if none exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := resetLicense(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "LICENSE"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "MIT License")
	})
}

func TestReplaceInTextFiles(t *testing.T) {
	t.Run("replaces patterns in markdown files", func(t *testing.T) {
		tmpDir := t.TempDir()

		readme := filepath.Join(tmpDir, "README.md")
		require.NoError(t, os.WriteFile(readme, []byte(
			"[![Build](https://github.com/peiman/ckeletin-go/actions)]\nProject: ckeletin-go\n",
		), 0600))

		replacements := []StringReplacement{
			{Old: "github.com/peiman/ckeletin-go", New: "github.com/me/myapp"},
			{Old: "peiman/ckeletin-go", New: "me/myapp"},
			{Old: "ckeletin-go", New: "myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, err := os.ReadFile(readme)
		require.NoError(t, err)
		got := string(content)
		assert.Contains(t, got, "me/myapp/actions")
		assert.Contains(t, got, "Project: myapp")
		assert.NotContains(t, got, "peiman")
		assert.NotContains(t, got, "ckeletin-go")
	})

	t.Run("replaces patterns in .yml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		yamlFile := filepath.Join(tmpDir, "config.yml")
		require.NoError(t, os.WriteFile(yamlFile, []byte(
			"reviewers:\n  - \"peiman\"\n",
		), 0600))

		replacements := []StringReplacement{
			{Old: "\"peiman\"", New: "\"me\""},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, err := os.ReadFile(yamlFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "\"me\"")
		assert.NotContains(t, string(content), "\"peiman\"")
	})

	t.Run("replaces patterns in .yaml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		yamlFile := filepath.Join(tmpDir, "config.yaml")
		require.NoError(t, os.WriteFile(yamlFile, []byte("name: ckeletin-go\n"), 0600))

		replacements := []StringReplacement{
			{Old: "ckeletin-go", New: "myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, err := os.ReadFile(yamlFile)
		require.NoError(t, err)
		assert.Equal(t, "name: myapp\n", string(content))
	})

	t.Run("skips .git, vendor, dist, and .task directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		for _, dir := range []string{".git", "vendor", "dist", ".task"} {
			d := filepath.Join(tmpDir, dir)
			require.NoError(t, os.MkdirAll(d, 0750))
			require.NoError(t, os.WriteFile(filepath.Join(d, "config.yml"), []byte("ckeletin-go"), 0600))
		}

		replacements := []StringReplacement{
			{Old: "ckeletin-go", New: "myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify files were NOT modified
		for _, dir := range []string{".git", "vendor", "dist", ".task"} {
			content, _ := os.ReadFile(filepath.Join(tmpDir, dir, "config.yml"))
			assert.Equal(t, "ckeletin-go", string(content), "%s directory should be skipped", dir)
		}
	})

	t.Run("skips Go source files", func(t *testing.T) {
		tmpDir := t.TempDir()

		goFile := filepath.Join(tmpDir, "main.go")
		require.NoError(t, os.WriteFile(goFile, []byte("// ckeletin-go"), 0600))

		replacements := []StringReplacement{
			{Old: "ckeletin-go", New: "myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		content, _ := os.ReadFile(goFile)
		assert.Equal(t, "// ckeletin-go", string(content))
	})

	t.Run("applies replacements in order (most specific first)", func(t *testing.T) {
		tmpDir := t.TempDir()

		mdFile := filepath.Join(tmpDir, "test.md")
		require.NoError(t, os.WriteFile(mdFile, []byte(
			"url: github.com/peiman/ckeletin-go\nowner: peiman/ckeletin-go\nname: ckeletin-go\n",
		), 0600))

		replacements := []StringReplacement{
			{Old: "github.com/peiman/ckeletin-go", New: "github.com/me/myapp"},
			{Old: "peiman/ckeletin-go", New: "me/myapp"},
			{Old: "ckeletin-go", New: "myapp"},
		}

		_, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)

		content, err := os.ReadFile(mdFile)
		require.NoError(t, err)
		got := string(content)
		assert.Equal(t, "url: github.com/me/myapp\nowner: me/myapp\nname: myapp\n", got)
	})

	t.Run("no changes when no matches", func(t *testing.T) {
		tmpDir := t.TempDir()

		mdFile := filepath.Join(tmpDir, "test.md")
		require.NoError(t, os.WriteFile(mdFile, []byte("no matches here"), 0600))

		replacements := []StringReplacement{
			{Old: "ckeletin-go", New: "myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("handles nested directory structures", func(t *testing.T) {
		tmpDir := t.TempDir()

		nestedDir := filepath.Join(tmpDir, ".github", "ISSUE_TEMPLATE")
		require.NoError(t, os.MkdirAll(nestedDir, 0750))
		require.NoError(t, os.WriteFile(
			filepath.Join(nestedDir, "config.yml"),
			[]byte("url: https://github.com/peiman/ckeletin-go/discussions"),
			0600,
		))

		replacements := []StringReplacement{
			{Old: "peiman/ckeletin-go", New: "me/myapp"},
		}

		count, err := replaceInTextFiles(tmpDir, replacements)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, _ := os.ReadFile(filepath.Join(nestedDir, "config.yml"))
		assert.Contains(t, string(content), "me/myapp/discussions")
	})
}

func TestRemovePkgDirectory(t *testing.T) {
	t.Run("removes pkg directory and nested contents", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a fake pkg/ structure with nested dirs
		pkgDir := filepath.Join(tmpDir, "pkg")
		checkmateDir := filepath.Join(pkgDir, "checkmate")
		demoDir := filepath.Join(pkgDir, "checkmate", "demo")
		require.NoError(t, os.MkdirAll(demoDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(checkmateDir, "checkmate.go"), []byte("package checkmate"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(demoDir, "main.go"), []byte("package main"), 0600))

		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)

		_, err = os.Stat(pkgDir)
		assert.True(t, os.IsNotExist(err), "pkg/ directory should be removed")
	})

	t.Run("no error when pkg directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)
	})

	t.Run("preserves other directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "pkg", "checkmate"), 0750))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "check"), 0750))
		internalFile := filepath.Join(tmpDir, "internal", "check", "check.go")
		require.NoError(t, os.WriteFile(internalFile, []byte("package check"), 0600))

		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, "internal", "check", "check.go"))
		assert.NoError(t, err, "internal/ files should be preserved")

		_, err = os.Stat(filepath.Join(tmpDir, "pkg"))
		assert.True(t, os.IsNotExist(err), "pkg/ should be removed")
	})
}

func TestRemoveFrameworkOnlyArtifacts(t *testing.T) {
	t.Run("removes conformance tests and mapping", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create framework-only artifacts
		conformDir := filepath.Join(tmpDir, "test", "conformance")
		require.NoError(t, os.MkdirAll(conformDir, 0750))
		require.NoError(t, os.WriteFile(
			filepath.Join(conformDir, "violation_test.go"),
			[]byte("package conformance"), 0600))

		require.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, "conformance-mapping.yaml"),
			[]byte("spec_version: 0.3.0"), 0600))

		scaffoldTest := filepath.Join(tmpDir, "test", "integration")
		require.NoError(t, os.MkdirAll(scaffoldTest, 0750))
		require.NoError(t, os.WriteFile(
			filepath.Join(scaffoldTest, "scaffold_init_test.go"),
			[]byte("package integration"), 0600))

		// Create a file that should survive
		require.NoError(t, os.WriteFile(
			filepath.Join(scaffoldTest, "other_test.go"),
			[]byte("package integration"), 0600))

		err := removeFrameworkOnlyArtifacts(tmpDir)
		assert.NoError(t, err)

		// Framework artifacts removed
		_, err = os.Stat(filepath.Join(tmpDir, "test", "conformance"))
		assert.True(t, os.IsNotExist(err), "test/conformance/ should be removed")

		_, err = os.Stat(filepath.Join(tmpDir, "conformance-mapping.yaml"))
		assert.True(t, os.IsNotExist(err), "conformance-mapping.yaml should be removed")

		_, err = os.Stat(filepath.Join(scaffoldTest, "scaffold_init_test.go"))
		assert.True(t, os.IsNotExist(err), "scaffold_init_test.go should be removed")

		// Other files preserved
		_, err = os.Stat(filepath.Join(scaffoldTest, "other_test.go"))
		assert.NoError(t, err, "other integration tests should be preserved")
	})

	t.Run("no error when artifacts do not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := removeFrameworkOnlyArtifacts(tmpDir)
		assert.NoError(t, err)
	})
}

func TestCleanArchLintConfig(t *testing.T) {
	t.Run("removes public component section", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".go-arch-lint.yml")

		content := `components:
  business:
    in:
      - internal/ping/**

  # -------------------------------------------------------------------------
  # PUBLIC PACKAGES: Standalone Reusable Libraries
  # -------------------------------------------------------------------------
  # pkg/ contains standalone packages that:
  # - Do NOT import from internal/ (enforced by validate-package-organization.sh)
  # - Can be imported by external Go projects
  # - Are independent of the CLI architecture
  #
  # See ADR-010 for guidance on when to use pkg/
  public:
    in:
      - pkg/**

vendors:
  cobra:
    in: github.com/spf13/cobra

commonComponents:
  - infrastructure
  - public  # Public packages can be used by any layer

deps:
  commands:
    mayDependOn:
      - business

  public:
    anyVendorDeps: true
`
		require.NoError(t, os.WriteFile(configPath, []byte(content), 0600))

		err := cleanArchLintConfig(tmpDir, "")
		require.NoError(t, err)

		result, err := os.ReadFile(configPath)
		require.NoError(t, err)
		got := string(result)

		// public component section removed
		assert.NotContains(t, got, "public:")
		assert.NotContains(t, got, "pkg/**")
		assert.NotContains(t, got, "anyVendorDeps")

		// public removed from commonComponents
		assert.NotContains(t, got, "- public")

		// Comment block removed
		assert.NotContains(t, got, "PUBLIC PACKAGES")
		assert.NotContains(t, got, "ADR-010")

		// Other content preserved
		assert.Contains(t, got, "business:")
		assert.Contains(t, got, "internal/ping/**")
		assert.Contains(t, got, "commonComponents:")
		assert.Contains(t, got, "- infrastructure")
		assert.Contains(t, got, "commands:")
		assert.Contains(t, got, "cobra:")
	})

	t.Run("no error when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := cleanArchLintConfig(tmpDir, "")
		assert.NoError(t, err)
	})

	t.Run("no error when file has no public section", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".go-arch-lint.yml")
		content := "components:\n  business:\n    in:\n      - internal/ping/**\n"
		require.NoError(t, os.WriteFile(configPath, []byte(content), 0600))

		err := cleanArchLintConfig(tmpDir, "")
		assert.NoError(t, err)

		result, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(result), "business:")
	})
}

func TestReplaceNameInGoFiles(t *testing.T) {
	t.Run("replaces old name in string literals and comments", func(t *testing.T) {
		tmpDir := t.TempDir()

		goFile := filepath.Join(tmpDir, "root.go")
		require.NoError(t, os.WriteFile(goFile, []byte(
			"package cmd\n\nvar binaryName = \"ckeletin-go\"\nvar logPath = \"./logs/ckeletin-go.log\"\n",
		), 0600))

		count, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, err := os.ReadFile(goFile)
		require.NoError(t, err)
		got := string(content)
		assert.Contains(t, got, "\"myapp\"")
		assert.Contains(t, got, "myapp.log")
		assert.NotContains(t, got, "ckeletin-go")
	})

	t.Run("skips .git vendor dist .task directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		for _, dir := range []string{".git", "vendor", "dist", ".task"} {
			d := filepath.Join(tmpDir, dir)
			require.NoError(t, os.MkdirAll(d, 0750))
			require.NoError(t, os.WriteFile(filepath.Join(d, "main.go"), []byte("// ckeletin-go"), 0600))
		}

		count, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("skips non-Go files", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("ckeletin-go"), 0600))

		count, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		content, _ := os.ReadFile(filepath.Join(tmpDir, "README.md"))
		assert.Equal(t, "ckeletin-go", string(content))
	})

	t.Run("preserves import statements", func(t *testing.T) {
		tmpDir := t.TempDir()

		goFile := filepath.Join(tmpDir, "executor.go")
		require.NoError(t, os.WriteFile(goFile, []byte(`package check

import (
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

var fallback = "ckeletin-go"
`), 0600))

		count, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		content, err := os.ReadFile(goFile)
		require.NoError(t, err)
		got := string(content)
		// Import line preserved (pkg/ import should NOT be modified)
		assert.Contains(t, got, "github.com/peiman/ckeletin-go/pkg/checkmate")
		// String literal replaced
		assert.Contains(t, got, `"myapp"`)
		assert.NotContains(t, got, `"ckeletin-go"`)
	})

	t.Run("preserves single-line imports", func(t *testing.T) {
		tmpDir := t.TempDir()

		goFile := filepath.Join(tmpDir, "simple.go")
		require.NoError(t, os.WriteFile(goFile, []byte(`package main

import "github.com/peiman/ckeletin-go/pkg/checkmate"

var name = "ckeletin-go"
`), 0600))

		_, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)

		content, err := os.ReadFile(goFile)
		require.NoError(t, err)
		got := string(content)
		assert.Contains(t, got, "github.com/peiman/ckeletin-go/pkg/checkmate")
		assert.Contains(t, got, `"myapp"`)
	})

	t.Run("no changes when name not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0600))

		count, err := replaceNameInGoFiles(tmpDir, "ckeletin-go", "myapp")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
