package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/catalog"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findGlobal(flags []catalog.Flag, long string) *catalog.Flag {
	for i := range flags {
		if flags[i].Long == long {
			return &flags[i]
		}
	}
	return nil
}

// TestBuildCatalog_RealCommandTree derives the catalog from the actual RootCmd —
// the same tree cobra parses — so the anti-drift property is exercised, not just
// argued: every command the parser knows is in the catalog.
func TestBuildCatalog_RealCommandTree(t *testing.T) {
	cat := buildCatalog(RootCmd)

	require.NotEmpty(t, cat.Name)

	names := map[string]bool{}
	for _, c := range cat.Commands {
		names[c.Name] = true
	}
	// Self-referential: the catalog command appears in its own catalog.
	assert.True(t, names["catalog"], "catalog should be self-referential")
	assert.True(t, names["config"], "config should be present")
	assert.True(t, names["completion"], "completion should be present")
	assert.False(t, names["help"], "cobra's auto help command must be excluded")

	// Recursion: config has the validate subcommand.
	for _, c := range cat.Commands {
		if c.Name == "config" {
			sub := map[string]bool{}
			for _, s := range c.Commands {
				sub[s.Name] = true
			}
			assert.True(t, sub["validate"], "config.validate should be walked recursively")
		}
	}

	// --output is a global, value-taking flag; help/version are excluded.
	out := findGlobal(cat.GlobalFlags, "output")
	require.NotNil(t, out, "--output should be a global flag")
	assert.True(t, out.TakesValue)
	assert.Nil(t, findGlobal(cat.GlobalFlags, "help"))
	assert.Nil(t, findGlobal(cat.GlobalFlags, "version"))

	// A boolean switch global → takes_value false.
	sw := findGlobal(cat.GlobalFlags, "log-file-enabled")
	require.NotNil(t, sw, "--log-file-enabled should be a global switch")
	assert.False(t, sw.TakesValue)
}

func TestCatalogFlags_RequiredTakesValueAndDefault(t *testing.T) {
	fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
	fs.StringP("name", "n", "", "the name")
	fs.Bool("force", false, "force it")
	fs.String("mode", "fast", "the mode")
	require.NoError(t, cobra.MarkFlagRequired(fs, "name"))

	byName := map[string]catalog.Flag{}
	for _, f := range catalogFlags(fs) {
		byName[f.Long] = f
	}

	// Required flag (via the BashCompOneRequiredFlag annotation), value-taking.
	assert.True(t, byName["name"].Required)
	assert.True(t, byName["name"].TakesValue)
	assert.Equal(t, "n", byName["name"].Short)

	// Boolean switch: not required, no value, false-default suppressed.
	assert.False(t, byName["force"].Required)
	assert.False(t, byName["force"].TakesValue)
	assert.Empty(t, byName["force"].Default, "a switch's false default is suppressed")

	// Value flag with a real default.
	assert.True(t, byName["mode"].TakesValue)
	assert.Equal(t, "fast", byName["mode"].Default)
}

func TestIncludeInCatalog_ExcludesHelpAndHidden(t *testing.T) {
	assert.False(t, includeInCatalog(&cobra.Command{Use: "help"}))
	assert.False(t, includeInCatalog(&cobra.Command{Use: "secret", Hidden: true,
		Run: func(*cobra.Command, []string) {}}))
	assert.True(t, includeInCatalog(&cobra.Command{Use: "ping",
		Run: func(*cobra.Command, []string) {}}))
}

func TestRunCatalog_JSONAndText(t *testing.T) {
	orig := output.OutputMode()
	defer output.SetOutputMode(orig)

	// JSON mode: a single success envelope whose data is the catalog.
	output.SetOutputMode("json")
	var jbuf bytes.Buffer
	catalogCmd.SetOut(&jbuf)
	require.NoError(t, runCatalog(catalogCmd, nil))

	var env map[string]any
	require.NoError(t, json.Unmarshal(jbuf.Bytes(), &env))
	assert.Equal(t, "success", env["status"])
	data, ok := env["data"].(map[string]any)
	require.True(t, ok, "envelope data should be the catalog object")
	assert.NotEmpty(t, data["name"])
	assert.Contains(t, data, "commands")

	// Text mode: human-readable listing.
	output.SetOutputMode("text")
	var tbuf bytes.Buffer
	catalogCmd.SetOut(&tbuf)
	require.NoError(t, runCatalog(catalogCmd, nil))
	assert.Contains(t, tbuf.String(), "Commands:")
}
