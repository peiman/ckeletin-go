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
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to bash when no shell is given (preserves prior behavior).
		shell := "bash"
		if len(args) > 0 {
			shell = args[0]
		}
		out := cmd.OutOrStdout()
		switch shell {
		case "bash":
			return cmd.Root().GenBashCompletion(out)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		default:
			return fmt.Errorf("unsupported shell %q (want bash, zsh, fish, or powershell)", shell)
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
