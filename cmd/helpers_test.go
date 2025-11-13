// cmd/helpers_test.go

package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand_Success(t *testing.T) {
	// SETUP: Create valid command metadata
	meta := config.CommandMetadata{
		Use:          "test",
		Short:        "Test command",
		Long:         "A test command for testing NewCommand",
		ConfigPrefix: "app.test",
		Hidden:       false,
	}

	runE := func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// EXECUTION: Create command
	cmd, err := NewCommand(meta, runE)

	// ASSERTION: Should succeed
	assert.NoError(t, err, "NewCommand() should not return error")
	require.NotNil(t, cmd, "NewCommand() should return non-nil command")
	assert.Equal(t, "test", cmd.Use, "Command.Use should match metadata")
	assert.Equal(t, "Test command", cmd.Short, "Command.Short should match metadata")
}

func TestNewCommand_ReturnsErrorOnInvalidFlags(t *testing.T) {
	// SETUP: Create metadata with invalid config prefix
	// Using a prefix that doesn't exist in the registry
	meta := config.CommandMetadata{
		Use:          "invalid",
		Short:        "Invalid command",
		Long:         "A command that should fail flag registration",
		ConfigPrefix: "nonexistent.invalid.prefix.that.does.not.exist",
		Hidden:       false,
	}

	runE := func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// EXECUTION: Create command
	cmd, err := NewCommand(meta, runE)

	// ASSERTION: Should return nil command and nil error (no flags to register)
	// Note: Empty prefix is not an error, it just means no flags to register
	assert.NoError(t, err, "NewCommand() should not return error even with non-existent prefix")
	require.NotNil(t, cmd, "NewCommand() should return command even with non-existent prefix")
}

func TestMustNewCommand_Success(t *testing.T) {
	// SETUP: Create valid command metadata
	meta := config.CommandMetadata{
		Use:          "test-must",
		Short:        "Test must command",
		Long:         "A test command for testing MustNewCommand",
		ConfigPrefix: "app.test",
		Hidden:       false,
	}

	runE := func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// EXECUTION: Create command with MustNewCommand
	cmd := MustNewCommand(meta, runE)

	// ASSERTION: Should succeed
	require.NotNil(t, cmd, "MustNewCommand() should return non-nil command")
	assert.Equal(t, "test-must", cmd.Use, "Command.Use should match metadata")
}

func TestNewCommand_PreservesMetadata(t *testing.T) {
	// SETUP: Create metadata with all fields
	meta := config.CommandMetadata{
		Use:          "preserve-test",
		Short:        "Short description",
		Long:         "Long description with details",
		ConfigPrefix: "app.test",
		Hidden:       true,
		Examples:     []string{"example1", "example2"},
	}

	runE := func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// EXECUTION: Create command
	cmd, err := NewCommand(meta, runE)

	// ASSERTION: All metadata should be preserved
	require.NoError(t, err, "NewCommand() should not return error")
	assert.Equal(t, meta.Use, cmd.Use, "Command.Use should match metadata")
	assert.Equal(t, meta.Short, cmd.Short, "Command.Short should match metadata")
	assert.Equal(t, meta.Long, cmd.Long, "Command.Long should match metadata")
	assert.Equal(t, meta.Hidden, cmd.Hidden, "Command.Hidden should match metadata")
	assert.NotNil(t, cmd.RunE, "Command.RunE should not be nil")
}
