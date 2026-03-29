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
