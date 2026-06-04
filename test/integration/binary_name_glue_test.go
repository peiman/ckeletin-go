// test/integration/binary_name_glue_test.go
//
// Guard tests for the binary-name -> config-dir "glue" — the chain that makes a
// scaffolded project read ~/.config/<its-name> instead of ~/.config/ckeletin-go:
//
//	Taskfile BINARY_NAME / goreleaser project_name
//	  -> ldflags  -X <module>/cmd.binaryName=<name>
//	  -> cmd.binaryName package var
//	  -> resolveXDGConfigDir() -> ~/.config/<name>
//
// Why these tests exist: every link is individually plausible-looking, yet two
// properties make a break SILENT — Go's linker treats `-X` to an unknown symbol
// as a no-op, and cmd/root.go's empty->"ckeletin-go" fallback masks a failed
// injection. So a rename of the `binaryName` var, a deleted `-X` line, or a
// dropped Taskfile passthrough would leave every existing unit test green while
// shipping binaries that read the wrong config directory.
//
// Two layers, both run in `task check` (no scaffold tag, no git/network/task
// dependency): a behavioral build-and-run proof (with a self-contained negative
// control) and a structural SSOT cross-check of every injection site.

package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// glueProbeName is a sentinel binary name chosen so it cannot collide with the
// upstream name ("ckeletin-go") or any real project name. The config directory a
// freshly-built binary resolves MUST contain this string — proving the ldflags
// injection actually landed in the live var.
const glueProbeName = "ckglueprobe7r"

// modulePath returns the current module path (e.g. github.com/peiman/ckeletin-go)
// read live, so the test never hardcodes a path that scaffolding rewrites.
func modulePath(t *testing.T, repoRoot string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	require.NoError(t, err, "go list -m failed")
	return strings.TrimSpace(string(out))
}

// buildWithBinaryNameSymbol builds the project binary into a temp dir, injecting
// the sentinel into the given ldflags SYMBOL path. A correct symbol lands in the
// live var; a wrong symbol is silently ignored by the linker — which is exactly
// the failure mode the negative control below pins down.
func buildWithBinaryNameSymbol(t *testing.T, repoRoot, symbol string) string {
	t.Helper()
	binName := glueProbeName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)
	// No shell quotes here (unlike the Taskfile's `-X '...'`): exec passes args
	// verbatim, so the single -ldflags arg carries `-X <symbol>=<value>`.
	ldflags := "-ldflags=-X " + symbol + "=" + glueProbeName
	cmd := exec.Command("go", "build", ldflags, "-o", binPath, "main.go")
	cmd.Dir = repoRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Run(), "build failed: %s", stderr.String())
	return binPath
}

// resolvedConfigFile runs `<bin> config validate --output json` under a controlled
// XDG_CONFIG_HOME and returns the config_file the binary resolved. A config file is
// planted ONLY under <xdg>/<plantedName>/, and the working directory is empty, so
// the returned path reveals which directory name the binary derived from its
// (injected) binaryName.
func resolvedConfigFile(t *testing.T, binPath, plantedName string) string {
	t.Helper()
	xdgHome := t.TempDir()
	cfgDir := filepath.Join(xdgHome, plantedName)
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"),
		[]byte("# ckeletin glue probe config\n"), 0o600))

	cmd := exec.Command(binPath, "config", "validate", "--output", "json")
	cmd.Dir = t.TempDir() // empty cwd, so the "." search path finds nothing
	cmd.Env = []string{
		"XDG_CONFIG_HOME=" + xdgHome,
		"HOME=" + t.TempDir(),
		"PATH=" + os.Getenv("PATH"),
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Exit code may be non-zero (e.g. warnings); the JSON envelope is emitted
	// regardless, and config_file is populated either way.
	_ = cmd.Run()

	payload := extractJSONObject(stdout.String())
	var env struct {
		Data struct {
			ConfigFile string `json:"config_file"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(payload), &env),
		"stdout was not a JSON envelope:\nstdout: %s\nstderr: %s", stdout.String(), stderr.String())
	return env.Data.ConfigFile
}

// extractJSONObject returns the outermost {...} from s, tolerating any stray
// surrounding output so the test asserts on the envelope, not on log noise.
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < start {
		return strings.TrimSpace(s)
	}
	return s[start : end+1]
}

// TestBinaryNameLdflagsDrivesConfigDir proves the runtime end of the glue: a binary
// built with `-X <module>/cmd.binaryName=<name>` resolves ~/.config/<name>. The
// negative control proves the assertion is meaningful and documents the silent-no-op
// hazard that makes a var rename undetectable to ordinary unit tests.
func TestBinaryNameLdflagsDrivesConfigDir(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	module := modulePath(t, repoRoot)

	t.Run("correct symbol injects the config directory name", func(t *testing.T) {
		bin := buildWithBinaryNameSymbol(t, repoRoot, module+"/cmd.binaryName")
		configFile := resolvedConfigFile(t, bin, glueProbeName)
		assert.Contains(t, configFile, glueProbeName,
			"the injected binaryName must drive the resolved config dir (got %q)", configFile)
	})

	// Negative control: Go's linker SILENTLY ignores `-X` to a non-existent symbol.
	// If this subtest ever started CONTAINING the probe name, it would mean the
	// build no longer depends on the exact symbol path — and the positive assertion
	// above would be a tautology. Keeping both pins the symbol path to live behavior.
	t.Run("non-existent symbol is silently ignored (falls back, not the probe)", func(t *testing.T) {
		bin := buildWithBinaryNameSymbol(t, repoRoot, module+"/cmd.binaryNameDOESNOTEXIST")
		configFile := resolvedConfigFile(t, bin, glueProbeName)
		assert.NotContains(t, configFile, glueProbeName,
			"an unknown ldflags symbol must NOT affect the config dir (got %q)", configFile)
	})
}

// TestBinaryNameInjectionSitesSSOT cross-checks the build-config end of the glue.
// The behavioral test builds via plain `go build`, so it cannot catch a break in
// the Taskfile passthrough or in goreleaser's separate RELEASE-binary injection.
// This structural test locks every injection site and the project_name<->BINARY_NAME
// invariant the .goreleaser.yml header explicitly requires. Matching is tolerant
// (no byte/whitespace equality) so reformatting does not cause false failures.
func TestBinaryNameInjectionSitesSSOT(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(repoRoot, rel))
		require.NoError(t, err, "reading %s", rel)
		return string(b)
	}
	frameworkTaskfile := read(".ckeletin/Taskfile.yml")
	projectTaskfile := read("Taskfile.yml")
	goreleaser := read(".goreleaser.yml")
	rootGo := read("cmd/root.go")

	// A) Framework LDFLAGS injects cmd.binaryName from the _BINARY_NAME var
	//    (not a hardcoded literal, which would ignore the project's BINARY_NAME).
	assert.Regexp(t, `/cmd\.binaryName=\{\{\s*\._BINARY_NAME\s*\}\}`, frameworkTaskfile,
		"`.ckeletin/Taskfile.yml` LDFLAGS must inject cmd.binaryName from {{._BINARY_NAME}}")

	// B) GoReleaser (the binaries users actually download) injects cmd.binaryName
	//    from .ProjectName. A separate path the behavioral build does NOT exercise.
	assert.Regexp(t, `/cmd\.binaryName=\{\{\s*\.ProjectName\s*\}\}`, goreleaser,
		"`.goreleaser.yml` ldflags must inject cmd.binaryName from {{ .ProjectName }}")

	// C) SSOT invariant (documented in the .goreleaser.yml header): goreleaser
	//    project_name MUST equal the project Taskfile BINARY_NAME, or release
	//    binaries and local builds resolve different config dirs.
	binaryName := firstSubmatch(t, projectTaskfile, `(?m)^\s*BINARY_NAME:\s*([A-Za-z0-9._-]+)\s*$`)
	projectName := firstSubmatch(t, goreleaser, `(?m)^\s*project_name:\s*([A-Za-z0-9._-]+)\s*$`)
	assert.Equal(t, binaryName, projectName,
		"`project_name` in .goreleaser.yml must match `BINARY_NAME` in Taskfile.yml (SSOT)")

	// D) The project Taskfile passes BINARY_NAME through to the included framework
	//    Taskfile; without this, _BINARY_NAME falls back to the framework default.
	assert.Regexp(t, `(?m)^\s*BINARY_NAME:\s*'?\{\{\s*\.BINARY_NAME\s*\}\}`, projectTaskfile,
		"Taskfile.yml includes.ckeletin.vars must pass BINARY_NAME through to the framework")

	// E) cmd/root.go declares the `binaryName` var that the `-X cmd.binaryName`
	//    flags target, and resolveXDGConfigDir actually derives the dir from it.
	//    A rename here is the silent-no-op hazard the behavioral negative control
	//    demonstrates; this catches it cheaply and structurally.
	assert.Regexp(t, `(?m)^\s*binaryName\s+=`, rootGo,
		"cmd/root.go must declare package var `binaryName` (the cmd.binaryName ldflags target)")
	xdgFn := regexp.MustCompile(`(?s)func resolveXDGConfigDir\(\).*?\n\}`).FindString(rootGo)
	assert.Contains(t, xdgFn, "binaryName",
		"resolveXDGConfigDir must derive the config dir from binaryName")
}

// firstSubmatch returns the first capture group of pattern in content, failing the
// test if the pattern does not match exactly one captured value.
func firstSubmatch(t *testing.T, content, pattern string) string {
	t.Helper()
	m := regexp.MustCompile(pattern).FindStringSubmatch(content)
	require.Len(t, m, 2, "pattern %q did not capture a value", pattern)
	return m[1]
}
