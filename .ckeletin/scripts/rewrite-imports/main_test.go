package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	oldModule = "github.com/peiman/ckeletin-go"
	newModule = "github.com/user/myapp"
)

func TestRewriteFile(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		preservePkg bool
		wantChanged bool
	}{
		{
			name: "basic import rewrite",
			input: `package cmd

import "github.com/peiman/ckeletin-go/internal/check"

func init() {}
`,
			expected: `package cmd

import "github.com/user/myapp/internal/check"

func init() {}
`,
			wantChanged: true,
		},
		{
			name: "multiple imports in same file",
			input: `package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/cmd"
	"github.com/peiman/ckeletin-go/internal/check"
	"github.com/peiman/ckeletin-go/internal/ping"
)

func main() { fmt.Println("hi") }
`,
			expected: `package cmd

import (
	"fmt"

	"github.com/user/myapp/cmd"
	"github.com/user/myapp/internal/check"
	"github.com/user/myapp/internal/ping"
)

func main() { fmt.Println("hi") }
`,
			wantChanged: true,
		},
		{
			name: "preserve pkg imports when flag is set",
			input: `package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/internal/check"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

func main() { fmt.Println("hi") }
`,
			expected: `package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/user/myapp/internal/check"
)

func main() { fmt.Println("hi") }
`,
			preservePkg: true,
			wantChanged: true,
		},
		{
			name: "rewrite pkg imports when preserve flag is not set",
			input: `package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

func main() { fmt.Println("hi") }
`,
			expected: `package cmd

import (
	"fmt"

	"github.com/user/myapp/pkg/checkmate"
)

func main() { fmt.Println("hi") }
`,
			preservePkg: false,
			wantChanged: true,
		},
		{
			name: "string constants with module path are NOT affected",
			input: `package cmd

import "fmt"

const upstream = "github.com/peiman/ckeletin-go"

func main() { fmt.Println(upstream) }
`,
			expected: `package cmd

import "fmt"

const upstream = "github.com/peiman/ckeletin-go"

func main() { fmt.Println(upstream) }
`,
			wantChanged: false,
		},
		{
			name: "comments with module path are NOT affected",
			input: `package cmd

// See github.com/peiman/ckeletin-go/internal/check for details
import "fmt"

func main() { fmt.Println("hi") }
`,
			expected: `package cmd

// See github.com/peiman/ckeletin-go/internal/check for details
import "fmt"

func main() { fmt.Println("hi") }
`,
			wantChanged: false,
		},
		{
			name: "file with no matching imports is unchanged",
			input: `package main

import (
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`,
			expected: `package main

import (
	"fmt"
	"os"
)

func main() { fmt.Println(os.Args) }
`,
			wantChanged: false,
		},
		{
			name: "named imports preserved",
			input: `package cmd

import (
	mycheck "github.com/peiman/ckeletin-go/internal/check"
)

func main() { mycheck.Run() }
`,
			expected: `package cmd

import (
	mycheck "github.com/user/myapp/internal/check"
)

func main() { mycheck.Run() }
`,
			wantChanged: true,
		},
		{
			name: "dot import preserved",
			input: `package cmd

import (
	. "github.com/peiman/ckeletin-go/internal/check"
)

func main() { Run() }
`,
			expected: `package cmd

import (
	. "github.com/user/myapp/internal/check"
)

func main() { Run() }
`,
			wantChanged: true,
		},
		{
			name: "bare module import (no subpackage)",
			input: `package cmd

import "github.com/peiman/ckeletin-go"

func main() {}
`,
			expected: `package cmd

import "github.com/user/myapp"

func main() {}
`,
			wantChanged: true,
		},
		{
			name: "does not match partial module prefix",
			input: `package cmd

import "github.com/peiman/ckeletin-go-extra/pkg/foo"

func main() {}
`,
			expected: `package cmd

import "github.com/peiman/ckeletin-go-extra/pkg/foo"

func main() {}
`,
			wantChanged: false,
		},
		{
			name: "string constant AND import together",
			input: `package cmd

import "github.com/peiman/ckeletin-go/internal/check"

const upstream = "github.com/peiman/ckeletin-go"

func main() { check.Run() }
`,
			expected: `package cmd

import "github.com/user/myapp/internal/check"

const upstream = "github.com/peiman/ckeletin-go"

func main() { check.Run() }
`,
			wantChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.go")
			require.NoError(t, os.WriteFile(filePath, []byte(tt.input), 0600))

			changed, err := rewriteFile(filePath, oldModule, newModule, tt.preservePkg)
			require.NoError(t, err)
			assert.Equal(t, tt.wantChanged, changed)

			result, err := os.ReadFile(filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestRewriteDir(t *testing.T) {
	t.Run("rewrites go files in directory tree", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested directory structure
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "cmd"), 0750))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "check"), 0750))

		goFile1 := filepath.Join(tmpDir, "cmd", "root.go")
		require.NoError(t, os.WriteFile(goFile1, []byte(`package cmd

import "github.com/peiman/ckeletin-go/internal/check"

func init() { check.Run() }
`), 0600))

		goFile2 := filepath.Join(tmpDir, "internal", "check", "check.go")
		require.NoError(t, os.WriteFile(goFile2, []byte(`package check

func Run() {}
`), 0600))

		count, err := rewriteDir(tmpDir, oldModule, newModule, false)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "only one file had matching imports")

		result, err := os.ReadFile(goFile1)
		require.NoError(t, err)
		assert.Contains(t, string(result), newModule+"/internal/check")
		assert.NotContains(t, string(result), oldModule)
	})

	t.Run("skips non-go files", func(t *testing.T) {
		tmpDir := t.TempDir()

		txtFile := filepath.Join(tmpDir, "README.md")
		content := "See github.com/peiman/ckeletin-go for details"
		require.NoError(t, os.WriteFile(txtFile, []byte(content), 0600))

		count, err := rewriteDir(tmpDir, oldModule, newModule, false)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		result, err := os.ReadFile(txtFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(result), "non-go file should be unchanged")
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		vendorDir := filepath.Join(tmpDir, "vendor", "github.com", "peiman", "ckeletin-go")
		require.NoError(t, os.MkdirAll(vendorDir, 0750))

		vendorFile := filepath.Join(vendorDir, "main.go")
		require.NoError(t, os.WriteFile(vendorFile, []byte(`package main

import "github.com/peiman/ckeletin-go/internal/check"

func main() {}
`), 0600))

		count, err := rewriteDir(tmpDir, oldModule, newModule, false)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "vendor files should be skipped")

		result, err := os.ReadFile(vendorFile)
		require.NoError(t, err)
		assert.Contains(t, string(result), oldModule, "vendor file should be unchanged")
	})

	t.Run("skips .git directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitDir := filepath.Join(tmpDir, ".git", "hooks")
		require.NoError(t, os.MkdirAll(gitDir, 0750))

		// .git shouldn't have .go files normally, but verify skip behavior
		gitFile := filepath.Join(gitDir, "pre-commit.go")
		require.NoError(t, os.WriteFile(gitFile, []byte(`package main

import "github.com/peiman/ckeletin-go/internal/check"

func main() {}
`), 0600))

		count, err := rewriteDir(tmpDir, oldModule, newModule, false)
		require.NoError(t, err)
		assert.Equal(t, 0, count, ".git files should be skipped")
	})

	t.Run("preserve-pkg mode skips pkg imports", func(t *testing.T) {
		tmpDir := t.TempDir()

		goFile := filepath.Join(tmpDir, "main.go")
		require.NoError(t, os.WriteFile(goFile, []byte(`package main

import (
	"github.com/peiman/ckeletin-go/internal/check"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

func main() {
	check.Run()
	checkmate.Run()
}
`), 0600))

		count, err := rewriteDir(tmpDir, oldModule, newModule, true)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		result, err := os.ReadFile(goFile)
		require.NoError(t, err)
		assert.Contains(t, string(result), newModule+"/internal/check")
		assert.Contains(t, string(result), oldModule+"/pkg/checkmate", "pkg/ import should be preserved")
	})

	t.Run("empty directory returns zero count", func(t *testing.T) {
		tmpDir := t.TempDir()

		count, err := rewriteDir(tmpDir, oldModule, newModule, false)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
