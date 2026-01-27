package marinatemd

import (
	"github.com/glueckkanja/marinatemd/internal/logger"
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
		logger.Log.Info("marinatemd", "version", Version, "commit", Commit, "built", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
