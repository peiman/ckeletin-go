// internal/logger/filtered_writer.go
//
// FilteredWriter provides level-based filtering for zerolog outputs.
// This allows different log levels to be sent to different destinations
// (e.g., INFO+ to console, DEBUG+ to file).

package logger

import (
	"io"

	"github.com/rs/zerolog"
)

// FilteredWriter wraps an io.Writer and filters log messages based on minimum level.
// It implements both io.Writer and zerolog.LevelWriter interfaces.
//
// Example usage:
//
//	consoleWriter := FilteredWriter{
//	    Writer:   os.Stdout,
//	    MinLevel: zerolog.InfoLevel,
//	}
//	fileWriter := FilteredWriter{
//	    Writer:   logFile,
//	    MinLevel: zerolog.DebugLevel,
//	}
//	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
type FilteredWriter struct {
	Writer   io.Writer
	MinLevel zerolog.Level
}

// WriteLevel implements zerolog.LevelWriter interface.
// It only writes the log message if the level meets or exceeds the minimum level.
func (w FilteredWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	// Filter: only write if log level >= minimum level
	// In zerolog, higher numeric values = higher severity
	// (TraceLevel=-1, DebugLevel=0, InfoLevel=1, WarnLevel=2, ErrorLevel=3, etc.)
	if level >= w.MinLevel {
		return w.Writer.Write(p)
	}

	// Discard the message but report success
	// This is important for zerolog to continue processing
	return len(p), nil
}

// Write implements io.Writer interface.
// This is called when the logger doesn't have level information.
// We pass through all writes since we can't determine the level.
func (w FilteredWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)
}
