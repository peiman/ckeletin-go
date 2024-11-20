// Package infrastructure handles all infrastructure-related operations.
package infrastructure

import "github.com/rs/zerolog"

// File system permissions
const (
	DirPerms  = 0o755 // Directory permissions (read/write for owner, read for others)
	FilePerms = 0o600 // File permissions (read/write for owner only)
)

// Configuration file defaults
const (
	DefaultConfigFileName = "ckeletin-go.json"
)

/*
Log Levels (from zerolog):
- zerolog.TraceLevel (-1) - Trace level logging, most verbose
- zerolog.DebugLevel (0)  - Debug level, for developers
- zerolog.InfoLevel (1)   - Info level, general operational entries (Default)
- zerolog.WarnLevel (2)   - Warning level, non-critical entries
- zerolog.ErrorLevel (3)  - Error level, high-priority entries
- zerolog.FatalLevel (4)  - Fatal level, application-fatal errors
- zerolog.PanicLevel (5)  - Panic level, severe errors
- zerolog.NoLevel (6)     - Disabled level
- zerolog.Disabled (7)    - Logging disabled completely

Configuration methods:
- Environment: CKELETIN_LOGLEVEL=1
- Config file: "logLevel": 1
- Command line: --log-level 1
*/

// Default log level (using zerolog's constant)
const DefaultLogLevel = zerolog.InfoLevel

// Ping command defaults
const (
	DefaultPingCount   = 1
	DefaultPingMessage = "pong"
)
