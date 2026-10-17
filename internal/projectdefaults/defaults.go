// Package projectdefaults is the project's default overlay for framework options.
//
// The TOML file is embedded in the binary. task ckeletin:update replaces
// .ckeletin/ and does not replace this file, so a project default survives
// a framework update. The registry remains the list of options that exist.
package projectdefaults

import _ "embed"

// TOML is the embedded project defaults file.
//
//go:embed ckeletin-project.toml
var TOML []byte
