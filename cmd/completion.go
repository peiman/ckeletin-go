// cmd/completion.go
// ckeletin:allow-custom-command

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate the autocompletion script for the specified shell",
	Long: fmt.Sprintf(`To load completions:

Bash:
  source <(%s completion bash)
Zsh:
  source <(%s completion zsh)
Fish:
  %s completion fish | source
PowerShell:
  %s completion powershell | Out-String | Invoke-Expression
`, binaryName, binaryName, binaryName, binaryName),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to bash if no args provided:
		return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
