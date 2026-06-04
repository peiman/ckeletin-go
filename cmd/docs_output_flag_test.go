// cmd/docs_output_flag_test.go

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsConfig_OutputFileNotOutput pins that the `docs config` command's output
// FILE flag is named --output-file, not --output. A local --output flag shadows
// the global --output (output FORMAT) flag, so `docs config --output json` would
// create a file literally named "json" instead of selecting JSON format.
func TestDocsConfig_OutputFileNotOutput(t *testing.T) {
	docsCmd := findSubcommand(RootCmd, "docs")
	require.NotNil(t, docsCmd, "docs command must exist")
	configSub := findSubcommand(docsCmd, "config")
	require.NotNil(t, configSub, "docs config command must exist")

	assert.NotNil(t, configSub.Flags().Lookup("output-file"),
		"docs config should expose the output file as --output-file")

	assert.Nil(t, configSub.LocalFlags().Lookup("output"),
		"docs config must NOT define a local --output flag — it shadows the global --output (format) flag")
}
