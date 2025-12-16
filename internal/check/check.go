// internal/check/check.go
//
// Shared types and check method implementations for the check command.

package check

import (
	"fmt"
	"strings"
)

// Category constants for filtering
const (
	CategoryEnvironment  = "environment"
	CategoryQuality      = "quality"
	CategoryArchitecture = "architecture"
	CategorySecurity     = "security"
	CategoryDependencies = "dependencies"
	CategoryTests        = "tests"
)

// AllCategories lists all valid category names
var AllCategories = []string{
	CategoryEnvironment,
	CategoryQuality,
	CategoryArchitecture,
	CategorySecurity,
	CategoryDependencies,
	CategoryTests,
}

// Config holds configuration for the check command
type Config struct {
	FailFast   bool
	Verbose    bool
	Parallel   bool
	Categories []string // Filter to specific categories (empty = all)
	ShowTiming bool     // Show duration for each check
}

// ValidateCategories checks if all provided categories are valid
func ValidateCategories(categories []string) error {
	validSet := make(map[string]bool)
	for _, c := range AllCategories {
		validSet[c] = true
	}
	var invalid []string
	for _, c := range categories {
		if !validSet[strings.ToLower(c)] {
			invalid = append(invalid, c)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid categories: %s (valid: %s)",
			strings.Join(invalid, ", "),
			strings.Join(AllCategories, ", "))
	}
	return nil
}

// checkMethods provides access to individual check implementations.
// Used internally by the Executor.
type checkMethods struct {
	cfg        Config
	coverage   float64           // Code coverage percentage from test run
	onCoverage func(cov float64) // Callback to send coverage to TUI
}
