package marinatemd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set via build flags.
	Version = "dev"
	// Commit is set via build flags.
	Commit = "none"
	// BuildDate is set via build flags.
	BuildDate = "unknown"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of marinatemd",
	Long:  `All software has versions. This is marinatemd's.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("marinatemd %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
