// internal/config/validator/formatter.go
//
// Formatting and presentation logic for validation results

package validator

import (
	"fmt"
	"io"
)

// FormatResult formats validation results for display
//
//nolint:errcheck // Display function - errors writing to output are not actionable
func FormatResult(result *Result, writer io.Writer) {
	// Print header
	fmt.Fprintf(writer, "Validating: %s\n\n", result.ConfigFile)

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Fprintf(writer, "❌ Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Fprintf(writer, "  %d. %v\n", i+1, err)
		}
		fmt.Fprintln(writer)
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Fprintf(writer, "⚠️  Warnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Fprintf(writer, "  %d. %s\n", i+1, warning)
		}
		fmt.Fprintln(writer)
	}

	// Print summary
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Fprintln(writer, "✅ Configuration is valid!")
	} else if result.Valid && len(result.Warnings) > 0 {
		fmt.Fprintln(writer, "✅ Configuration is valid (with warnings)")
	} else {
		fmt.Fprintln(writer, "❌ Configuration is invalid")
	}
}

// ExitCodeForResult determines the appropriate exit code for validation results
// Returns nil for success (exit 0), error for failure/warnings (exit 1)
func ExitCodeForResult(result *Result) error {
	if result.Valid && len(result.Warnings) == 0 {
		// Exit code 0: Valid with no warnings
		return nil
	} else if result.Valid && len(result.Warnings) > 0 {
		// Exit code 1: Valid but with warnings
		return fmt.Errorf("validation completed with warnings")
	} else {
		// Exit code 1: Invalid (has errors)
		return fmt.Errorf("validation failed")
	}
}
