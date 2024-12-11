// cmd/completion.go

package cmd

import (
	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate the autocompletion script for the specified shell",
	Long: `To load completions:

Bash:
  source <(ckeletin-go completion bash)
Zsh:
  source <(ckeletin-go completion zsh)
Fish:
  ckeletin-go completion fish | source
PowerShell:
  ckeletin-go completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to bash if no args provided:
		return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
