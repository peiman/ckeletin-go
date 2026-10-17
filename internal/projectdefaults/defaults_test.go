package projectdefaults

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDefaultsFileOutsideFramework(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(file)
	slash := filepath.ToSlash(dir)
	assert.NotContains(t, slash, "/.ckeletin/")
	assert.NotEqual(t, ".ckeletin", filepath.Base(dir))

	path := filepath.Join(dir, "ckeletin-project.toml")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, body, TOML, "the binary must carry the file bytes, not a path read at runtime")
	assert.Contains(t, string(body), "[defaults]")
}

func TestRootAppliesEmbeddedDefaultsBeforeConfigFile(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "..", "cmd", "root.go")
	body, err := os.ReadFile(root)
	require.NoError(t, err)
	text := string(body)

	set := strings.Index(text, "config.SetDefaults()")
	apply := strings.Index(text, "config.ApplyProjectDefaultsTOML(projectdefaults.TOML)")
	validate := strings.Index(text, "config.ValidateAllConfigValues(")
	read := strings.Index(text, "viper.ReadInConfig()")
	require.NotEqual(t, -1, set)
	require.NotEqual(t, -1, apply)
	require.NotEqual(t, -1, validate)
	require.NotEqual(t, -1, read)
	assert.Less(t, set, apply, "project defaults apply after the registry defaults")
	assert.Less(t, apply, validate, "project defaults apply before default validation")
	assert.Less(t, apply, read, "project defaults apply before the config file is read")
}
