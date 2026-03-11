package dev

import (
	"fmt"
	"io"
)

// ConfigCommandFlags holds the flags for the config subcommand
type ConfigCommandFlags struct {
	List     bool
	Show     bool
	Export   string
	Validate bool
	Prefix   string
}

// RunConfigCommand executes the dev config subcommand logic.
// Returns nil if a flag was handled, or an error.
// Returns (false, nil) if no flags were specified (caller should show help).
func RunConfigCommand(w io.Writer, flags ConfigCommandFlags) (handled bool, err error) {
	ci := NewConfigInspector()

	if flags.List {
		_, _ = fmt.Fprintln(w, ci.FormatAsTable())
		return true, nil
	}

	if flags.Show {
		_, _ = fmt.Fprintln(w, ci.FormatEffectiveAsTable())
		return true, nil
	}

	if flags.Export != "" {
		switch flags.Export {
		case "json":
			jsonStr, err := ci.ExportToJSON(true)
			if err != nil {
				return true, fmt.Errorf("failed to export to JSON: %w", err)
			}
			_, _ = fmt.Fprintln(w, jsonStr)
			return true, nil
		default:
			return true, fmt.Errorf("unsupported export format: %s (supported: json)", flags.Export)
		}
	}

	if flags.Validate {
		errors := ci.ValidateConfig()
		if len(errors) > 0 {
			_, _ = fmt.Fprintln(w, "❌ Configuration validation failed:")
			for i, err := range errors {
				_, _ = fmt.Fprintf(w, "  %d. %v\n", i+1, err)
			}
			return true, fmt.Errorf("validation found %d error(s)", len(errors))
		}
		_, _ = fmt.Fprintln(w, "✅ Configuration is valid")
		return true, nil
	}

	if flags.Prefix != "" {
		matches := ci.GetConfigByPrefix(flags.Prefix)
		if len(matches) == 0 {
			_, _ = fmt.Fprintf(w, "No configuration options found with prefix: %s\n", flags.Prefix)
			return true, nil
		}
		_, _ = fmt.Fprintf(w, "Configuration options with prefix '%s':\n\n", flags.Prefix)
		for _, opt := range matches {
			_, _ = fmt.Fprintf(w, "  %s (%s): %s\n", opt.Key, opt.Type, opt.Description)
		}
		return true, nil
	}

	return false, nil
}
