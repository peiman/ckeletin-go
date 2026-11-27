package logger

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHoldConsole_WritesLinesWhenTheDisplayEnds(t *testing.T) {
	console := &bytes.Buffer{}
	viper.Set(config.KeyAppLogFileEnabled, false)
	viper.Set(config.KeyAppLogConsoleLevel, "info")
	viper.Set(config.KeyAppLogColorEnabled, "false")
	viper.Set(config.KeyAppOutputFormat, "text")
	t.Cleanup(func() {
		ReleaseConsole()
		Cleanup()
		viper.Set(config.KeyAppLogConsoleLevel, "")
		viper.Set(config.KeyAppOutputFormat, "")
	})
	require.NoError(t, Init(console))
	HoldConsole()
	logInfo("one")
	logInfo("two")
	assert.NotContains(t, console.String(), "one")
	ReleaseConsole()
	text := console.String()
	assert.Less(t, strings.Index(text, "one"), strings.Index(text, "two"))
}

func TestHoldConsole_JSONModeStaysSilent(t *testing.T) {
	console := &bytes.Buffer{}
	dir := t.TempDir()
	viper.Set(config.KeyAppLogFileEnabled, true)
	viper.Set(config.KeyAppLogFilePath, filepath.Join(dir, "app.log"))
	viper.Set(config.KeyAppLogFileLevel, "info")
	viper.Set(config.KeyAppLogConsoleLevel, "info")
	viper.Set(config.KeyAppOutputFormat, "json")
	t.Cleanup(func() {
		ReleaseConsole()
		Cleanup()
		viper.Set(config.KeyAppLogFileEnabled, false)
		viper.Set(config.KeyAppLogConsoleLevel, "")
		viper.Set(config.KeyAppOutputFormat, "")
	})
	require.NoError(t, Init(console))
	HoldConsole()
	logInfo("kept-in-file")
	ReleaseConsole()
	assert.Empty(t, console.String())
	body, err := os.ReadFile(filepath.Join(dir, "app.log"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "kept-in-file")
}

func logInfo(message string) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg(message)
}

func TestAppendLog_TwoProcessesLoseNothing(t *testing.T) {
	if os.Getenv("CKELETIN_APPEND_CHILD") == "1" {
		path := os.Getenv("CKELETIN_APPEND_PATH")
		id := os.Getenv("CKELETIN_APPEND_ID")
		writer, err := openAppendLog(path, 1<<20, 3, 1, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open: %v\n", err)
			os.Exit(2)
		}
		for i := 0; i < 40; i++ {
			if _, err := fmt.Fprintf(writer, "id=%s n=%02d\n", id, i); err != nil {
				fmt.Fprintf(os.Stderr, "write: %v\n", err)
				os.Exit(3)
			}
		}
		_ = writer.Close()
		os.Exit(0)
	}

	path := filepath.Join(t.TempDir(), "app.log")
	start := func(id string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestAppendLog_TwoProcesses", "-test.v")
		cmd.Env = append(os.Environ(),
			"CKELETIN_APPEND_CHILD=1",
			"CKELETIN_APPEND_PATH="+path,
			"CKELETIN_APPEND_ID="+id,
		)
		return cmd
	}
	first := start("a")
	second := start("b")
	require.NoError(t, first.Start())
	require.NoError(t, second.Start())
	require.NoError(t, first.Wait())
	require.NoError(t, second.Wait())

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	assert.Len(t, lines, 80)
	seen := map[string]int{}
	for _, line := range lines {
		seen[line]++
		assert.True(t, strings.HasPrefix(line, "id="), line)
	}
	assert.Len(t, seen, 80)
}
