package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
	Commit  = "none"
	Date    = "unknown"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ckeletin-go.",
	Long:  `All software has versions. This is ckeletin-go's.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("ckeletin-go v%s (built on %s, commit %s)\n", Version, Date, Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
