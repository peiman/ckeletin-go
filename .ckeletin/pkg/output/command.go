package output

import (
	"strings"

	"github.com/spf13/cobra"
)

// currentCmd is the command whose PersistentPreRunE last ran.
// Result and Success use it when the caller does not pass a command.
var currentCmd *cobra.Command

// CommandID is the envelope command identifier: the full path without the
// program name, words separated by a single space ("check model"). A top-level
// command is its own name ("version"). The root command alone is empty, which
// the error path marshals as JSON null.
func CommandID(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	full := strings.TrimSpace(cmd.CommandPath())
	rootName := ""
	if cmd.Root() != nil {
		rootName = cmd.Root().Name()
	}
	if rootName == "" {
		return full
	}
	if full == rootName {
		return ""
	}
	prefix := rootName + " "
	if strings.HasPrefix(full, prefix) {
		return strings.TrimSpace(full[len(prefix):])
	}
	return full
}

// SetCommandFrom stores the command id from cmd unless SetCommandName overrode it.
func SetCommandFrom(cmd *cobra.Command) {
	currentCmd = cmd
	if commandNameSet {
		return
	}
	currentCommandName = CommandID(cmd)
}

func commandLabel(cmd *cobra.Command) string {
	if commandNameSet || currentCommandName != "" {
		return currentCommandName
	}
	return CommandID(cmd)
}
