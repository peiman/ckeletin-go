// Package catalog defines the machine-readable command catalog schema
// (CKSPEC-AGENT-006).
//
// These types ARE the catalog schema: emitting them as the data of an OUT-002
// success envelope is how a ckeletin CLI self-reports its command surface as
// structured data, so an agent discovers capabilities without parsing human
// --help text or trusting hand-written AGENTS.md prose.
//
// The schema is the cross-implementation contract agreed with ckeletin-rust on
// the spec issue (CKSPEC-AGENT-006): a required core every implementation
// derives losslessly (so one parser works across both), plus optional fields
// each emits where its CLI framework exposes them (omitted, never hand-filled).
//
// Required core:
//   - command:   name, description, flags, commands (recursive)
//   - flag:      long, required, takes_value
//   - top level: name, description, global_flags (listed once), commands
//
// Optional (cobra fills what it can derive): flag short, description, default.
// cobra cannot structurally derive enumerated flag values, so possible_values is
// always omitted on the go side — the documented cobra/clap asymmetry.
//
// Defining the types in the framework package (synced downstream by
// `task ckeletin:update`) makes the schema a single shared type: a derived
// project cannot emit a wrong-shaped catalog. The cobra -> Catalog walk lives in
// cmd/ (cobra is the commands layer only, CKSPEC-ARCH-003) — the only
// implementation-specific part, mirroring ckeletin-rust's clap walk in its cli
// crate.
package catalog

import (
	"fmt"
	"strings"
)

// Flag is a single flag in the catalog.
type Flag struct {
	// Long is the long form, without "--" (required core).
	Long string `json:"long"`
	// Required is whether the flag must be supplied (required core).
	Required bool `json:"required"`
	// TakesValue is whether the flag consumes a value ("--x value") vs. a boolean
	// switch (required core — the normalized intersection both clap and cobra
	// derive).
	TakesValue bool `json:"takes_value"`
	// Short is the short form, without "-" (optional).
	Short string `json:"short,omitempty"`
	// Description is the help text (optional).
	Description string `json:"description,omitempty"`
	// Default is the default value, if any (optional).
	Default string `json:"default,omitempty"`
	// PossibleValues lists allowed values for enumerated flags (optional). Always
	// empty on the go side: cobra does not expose enumerated flag values as
	// structured data.
	PossibleValues []string `json:"possible_values,omitempty"`
}

// Command is a command and its subcommands.
type Command struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Flags       []Flag    `json:"flags"`
	Commands    []Command `json:"commands"`
}

// Catalog is the whole command surface of a CLI — the data of an OUT-002 success
// envelope.
type Catalog struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// GlobalFlags apply to every command, listed once here (not duplicated into
	// each command's Flags).
	GlobalFlags []Flag    `json:"global_flags"`
	Commands    []Command `json:"commands"`
}

// String renders a human-readable listing for text output mode.
func (c Catalog) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s - %s\n", c.Name, c.Description)
	if len(c.GlobalFlags) > 0 {
		b.WriteString("\nGlobal flags:\n")
		for _, f := range c.GlobalFlags {
			writeFlag(&b, f)
		}
	}
	b.WriteString("\nCommands:\n")
	for _, cmd := range c.Commands {
		fmt.Fprintf(&b, "  %-12s %s\n", cmd.Name, cmd.Description)
	}
	return b.String()
}

func writeFlag(b *strings.Builder, f Flag) {
	value := ""
	if f.TakesValue {
		value = " <value>"
	}
	desc := ""
	if f.Description != "" {
		desc = "  " + f.Description
	}
	fmt.Fprintf(b, "  --%s%s%s\n", f.Long, value, desc)
}
