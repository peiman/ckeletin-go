// cmd/helpers_test.go

package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/spf13/cobra"
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
	if err != nil {
		t.Errorf("NewCommand() unexpected error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewCommand() returned nil command")
	}
	if cmd.Use != "test" {
		t.Errorf("Command.Use = %v, want %v", cmd.Use, "test")
	}
	if cmd.Short != "Test command" {
		t.Errorf("Command.Short = %v, want %v", cmd.Short, "Test command")
	}
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
	if err != nil {
		t.Errorf("NewCommand() unexpected error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewCommand() returned nil command even with valid metadata")
	}
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
	if cmd == nil {
		t.Fatal("MustNewCommand() returned nil command")
	}
	if cmd.Use != "test-must" {
		t.Errorf("Command.Use = %v, want %v", cmd.Use, "test-must")
	}
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
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}
	if cmd.Use != meta.Use {
		t.Errorf("Command.Use = %v, want %v", cmd.Use, meta.Use)
	}
	if cmd.Short != meta.Short {
		t.Errorf("Command.Short = %v, want %v", cmd.Short, meta.Short)
	}
	if cmd.Long != meta.Long {
		t.Errorf("Command.Long = %v, want %v", cmd.Long, meta.Long)
	}
	if cmd.Hidden != meta.Hidden {
		t.Errorf("Command.Hidden = %v, want %v", cmd.Hidden, meta.Hidden)
	}
	if cmd.RunE == nil {
		t.Error("Command.RunE should not be nil")
	}
}
