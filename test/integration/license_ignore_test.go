package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go-licenses exits when it cannot name a license. LICENSE_ALLOWED cannot
// fix that, because the module never gets a license name. .license-ignore
// lists those module paths; anything not listed must still fail the check.

const licenseClassifyStub = `#!/bin/sh
printf '%s\n' "$@" > "$LICENSE_ARGS_FILE"
for a in "$@"; do
  case "$a" in
    *github.com/segmentio/asm*) exit 0 ;;
  esac
done
echo "Did not find license for library github.com/segmentio/asm" >&2
exit 1
`

const licenseAlwaysOKStub = `#!/bin/sh
printf '%s\n' "$@" > "$LICENSE_ARGS_FILE"
exit 0
`

func runLicenseCheck(t *testing.T, dir, stub string) (string, string, error) {
	t.Helper()
	bin := t.TempDir()
	argsFile := filepath.Join(bin, "args.txt")
	writeExec(t, bin, "go", "#!/bin/sh\necho github.com/test/fixture\n")
	writeExec(t, bin, "go-licenses", stub)

	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	cmd := exec.Command("bash", filepath.Join(root, ".ckeletin", "scripts", "check-licenses-source.sh"))
	cmd.Dir = dir
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "PATH=") || strings.HasPrefix(kv, "LICENSE_ARGS_FILE=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"LICENSE_ARGS_FILE="+argsFile,
	)

	out, err := cmd.CombinedOutput()
	args, readErr := os.ReadFile(argsFile)
	require.NoError(t, readErr, "go-licenses stub was not invoked\n%s", out)
	return string(out), string(args), err
}

func TestLicenseIgnore_ListedModuleIsPassedToGoLicenses(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".license-ignore"), []byte(`
# MIT-0. go-licenses cannot classify this module.
github.com/segmentio/asm MIT-0 unclassified

github.com/other/kept still checked
`), 0o644))

	out, args, err := runLicenseCheck(t, dir, licenseClassifyStub)
	require.NoError(t, err, "a listed module must be ignored\n%s", out)
	assert.Contains(t, args, "--ignore=github.com/segmentio/asm")
	assert.Contains(t, args, "--ignore=github.com/other/kept")
	assert.NotContains(t, args, "MIT-0")
	assert.NotContains(t, args, "unclassified")
	assert.Contains(t, out, "All dependency licenses compliant")
}

func TestLicenseIgnore_UnlistedModuleStillFails(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".license-ignore"), []byte("github.com/other/kept still checked\n"), 0o644))

	out, args, err := runLicenseCheck(t, dir, licenseClassifyStub)
	require.Error(t, err, "a module that is not listed must still fail the check\n%s", out)
	assert.Contains(t, args, "--ignore=github.com/other/kept")
	assert.NotContains(t, args, "github.com/segmentio/asm")
	assert.Contains(t, out, "License compliance check failed")
}

func TestLicenseIgnore_AbsentFileStillChecksTheModuleItself(t *testing.T) {
	dir := t.TempDir()

	out, args, err := runLicenseCheck(t, dir, licenseAlwaysOKStub)
	require.NoError(t, err, "a missing .license-ignore must not change a clean check\n%s", out)
	assert.Contains(t, args, "--ignore=github.com/test/fixture")
	assert.NotContains(t, args, ".license-ignore")
}
