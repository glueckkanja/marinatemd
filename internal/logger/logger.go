package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

// Log is the global logger instance used throughout the application.
var Log *log.Logger

// Setup initializes the global logger with the specified verbosity level.
// By default (no flags), only warnings and errors are shown.
// With --verbose, informational messages are shown.
// With --debug, all messages including debug output are shown.
func Setup(verbose, debug bool) {
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: debug, // Only show timestamps in debug mode
		TimeFormat:      time.Kitchen,
	})

	// Set log level based on flags
	switch {
	case debug:
		Log.SetLevel(log.DebugLevel)
	case verbose:
		Log.SetLevel(log.InfoLevel)
	default:
		Log.SetLevel(log.WarnLevel)
	}
}

// init ensures a default logger is always available, even if Setup is not called.
func init() {
	Setup(false, false)
}
