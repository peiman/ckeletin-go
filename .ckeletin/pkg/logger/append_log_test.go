package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendLog_CloseBeforeOpenAndMissingBackup(t *testing.T) {
	var writer appendLog
	require.NoError(t, writer.Close())
	compressBackup(filepath.Join(t.TempDir(), "missing"))
}

type failWrite struct{}

func (failWrite) Write([]byte) (int, error) { return 0, os.ErrClosed }

func TestSingleKeyJSON_ReturnsTheWriterError(t *testing.T) {
	_, err := (singleKeyJSON{Writer: failWrite{}}).Write([]byte("{\"a\":1}\n"))
	require.Error(t, err)
}

func TestAppendLog_RejectsAFileAsTheDirectory(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "not-a-directory")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
	_, err := openAppendLog(filepath.Join(blocker, "app.log"), 1024, 1, 1, false)
	require.Error(t, err)
}

func TestAppendLog_RotatesAndPrunes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	writer, err := openAppendLog(path, 32, 1, 1, true)
	require.NoError(t, err)
	_, err = writer.Write([]byte(strings.Repeat("a", 40) + "\n"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("next\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	_, err = writer.Write([]byte("after-close\n"))
	require.NoError(t, err)

	matches, err := filepath.Glob(path + ".*")
	require.NoError(t, err)
	assert.NotEmpty(t, matches)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		gz, _ := filepath.Glob(path + ".*.gz")
		if len(gz) > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestAppendLog_DropsAnOldBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	old := path + ".20000101T000000.000000000Z"
	require.NoError(t, os.WriteFile(old, []byte("old\n"), 0o600))
	past := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(old, past, past))

	writer, err := openAppendLog(path, 8, 5, 1, false)
	require.NoError(t, err)
	_, err = writer.Write([]byte("0123456789\n"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("more\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	_, err = os.Stat(old)
	assert.True(t, os.IsNotExist(err))
}

func TestCollapseJSONKeys(t *testing.T) {
	raw := []byte("{\"component\":\"cli\",\"component\":\"engine\"}\n")
	out := collapseJSONKeys(raw)
	assert.Equal(t, 1, strings.Count(string(out), `"component"`))
	assert.Contains(t, string(out), "engine")
	assert.Equal(t, "not-json", string(collapseJSONKeys([]byte("not-json"))))
	assert.Equal(t, "{", string(collapseJSONKeys([]byte("{"))))
}

func TestConsoleHold_NilOut(t *testing.T) {
	var hold consoleHold
	n, err := hold.Write([]byte("x"))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	hold.held = true
	_, err = hold.Write([]byte("y"))
	assert.NoError(t, err)
	ReleaseConsole()
}
