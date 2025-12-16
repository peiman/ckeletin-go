// internal/check/check.go
//
// Check command executor using pkg/checkmate for beautiful output.

package check

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/rs/zerolog/log"
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

// Executor handles the execution of the check command
type Executor struct {
	cfg        Config
	printer    checkmate.PrinterInterface
	writer     io.Writer
	onCoverage func(float64) // Callback for coverage percentage
	coverage   float64       // Stored coverage for display
}

// NewExecutor creates a new check command executor
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	printer := checkmate.New(checkmate.WithWriter(writer))
	return &Executor{
		cfg:     cfg,
		printer: printer,
		writer:  writer,
	}
}

// NewExecutorWithPrinter creates an executor with a custom printer (for testing)
func NewExecutorWithPrinter(cfg Config, printer checkmate.PrinterInterface, writer io.Writer) *Executor {
	return &Executor{
		cfg:     cfg,
		printer: printer,
		writer:  writer,
	}
}

// shouldRunCategory returns true if the category should be executed
func (e *Executor) shouldRunCategory(category string) bool {
	if len(e.cfg.Categories) == 0 {
		return true // No filter = run all
	}
	for _, c := range e.cfg.Categories {
		if strings.EqualFold(c, category) {
			return true
		}
	}
	return false
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

// Execute runs all quality checks
func (e *Executor) Execute(ctx context.Context) error {
	log.Debug().
		Bool("fail_fast", e.cfg.FailFast).
		Bool("verbose", e.cfg.Verbose).
		Bool("parallel", e.cfg.Parallel).
		Strs("categories", e.cfg.Categories).
		Bool("show_timing", e.cfg.ShowTiming).
		Msg("Starting quality checks")

	// Run checks in categories
	// Each category can run in parallel internally, but categories run sequentially
	// to provide clear progress feedback

	var totalPassed, totalFailed, totalChecks int

	// Category 1: Development Environment
	if e.shouldRunCategory(CategoryEnvironment) {
		result := e.runCategory(ctx, "Development Environment", []checkDef{
			{"go-version", e.shellCheck("check-go-version.sh"), "Ensure Go version matches .go-version"},
			{"tools", e.shellCheck("install_tools.sh", "--check"), "Run: task setup"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
		if e.cfg.FailFast && result.Failed > 0 {
			return e.finalResult(totalPassed, totalFailed, totalChecks)
		}
	}

	// Category 2: Code Quality (native Go checks)
	if e.shouldRunCategory(CategoryQuality) {
		result := e.runCategory(ctx, "Code Quality", []checkDef{
			{"format", e.checkFormat, "Run: task format"},
			{"lint", e.checkLint, "Run: task lint"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
		if e.cfg.FailFast && result.Failed > 0 {
			return e.finalResult(totalPassed, totalFailed, totalChecks)
		}
	}

	// Category 3: Architecture Validation
	if e.shouldRunCategory(CategoryArchitecture) {
		result := e.runCategory(ctx, "Architecture Validation", []checkDef{
			{"defaults", e.shellCheck("check-defaults.sh"), "Use registry for SetDefault (ADR-002)"},
			{"commands", e.shellCheck("validate-command-patterns.sh"), "Keep commands ultra-thin (ADR-001)"},
			{"constants", e.shellCheck("check-constants.sh"), "Run: task generate:config:key-constants"},
			{"task-naming", e.shellCheck("validate-task-naming.sh"), "Follow ADR-000 naming convention"},
			{"architecture", e.shellCheck("validate-architecture.sh"), "Update ARCHITECTURE.md (ADR-008)"},
			{"layering", e.shellCheck("validate-layering.sh"), "Fix layer dependencies (ADR-009)"},
			{"package-org", e.shellCheck("validate-package-organization.sh"), "Follow package organization (ADR-010)"},
			{"config-consumption", e.shellCheck("validate-config-consumption.sh"), "Use type-safe config (ADR-002)"},
			{"output-patterns", e.shellCheck("validate-output-patterns.sh"), "Follow output patterns (ADR-012)"},
			{"security-patterns", e.shellCheck("validate-security-patterns.sh"), "Implement security patterns (ADR-004)"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
		if e.cfg.FailFast && result.Failed > 0 {
			return e.finalResult(totalPassed, totalFailed, totalChecks)
		}
	}

	// Category 4: Security Scanning
	if e.shouldRunCategory(CategorySecurity) {
		result := e.runCategory(ctx, "Security Scanning", []checkDef{
			{"secrets", e.shellCheck("check-secrets.sh"), "Remove hardcoded secrets"},
			{"sast", e.shellCheck("check-sast.sh"), "Fix SAST issues or update .semgrep.yml"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
		if e.cfg.FailFast && result.Failed > 0 {
			return e.finalResult(totalPassed, totalFailed, totalChecks)
		}
	}

	// Category 5: Dependencies
	if e.shouldRunCategory(CategoryDependencies) {
		result := e.runCategory(ctx, "Dependencies", []checkDef{
			{"deps", e.checkDeps, "Run: go mod tidy"},
			{"vuln", e.checkVuln, "Update vulnerable dependencies"},
			{"license-source", e.shellCheck("check-licenses-source.sh"), "Check dependency licenses"},
			{"license-binary", e.shellCheck("check-licenses-binary.sh"), "Check binary licenses"},
			{"sbom-vulns", e.shellCheck("check-sbom-vulns.sh"), "Fix SBOM vulnerabilities"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
		if e.cfg.FailFast && result.Failed > 0 {
			return e.finalResult(totalPassed, totalFailed, totalChecks)
		}
	}

	// Category 6: Tests (native Go)
	if e.shouldRunCategory(CategoryTests) {
		result := e.runCategory(ctx, "Tests", []checkDef{
			{"test", e.checkTest, "Fix failing tests"},
		})
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalChecks += result.Total
	}

	// Display coverage if available
	if e.coverage > 0 {
		e.printer.CheckInfo(fmt.Sprintf("Coverage: %.1f%%", e.coverage))
	}

	// Only log in verbose mode to avoid noise in CI output
	if e.cfg.Verbose {
		log.Info().
			Int("passed", totalPassed).
			Int("failed", totalFailed).
			Int("total", totalChecks).
			Float64("coverage", e.coverage).
			Msg("Check run completed")
	}

	return e.finalResult(totalPassed, totalFailed, totalChecks)
}

// checkDef defines a single check with name, function, and remediation.
type checkDef struct {
	name        string
	fn          func(ctx context.Context) error
	remediation string
}

// runCategory runs a category of checks and returns the result.
func (e *Executor) runCategory(ctx context.Context, category string, checks []checkDef) checkmate.RunResult {
	opts := []checkmate.RunnerOption{
		checkmate.WithCategory(category),
	}
	if e.cfg.FailFast {
		opts = append(opts, checkmate.WithFailFast())
	}
	if e.cfg.Parallel {
		opts = append(opts, checkmate.WithParallel())
	}
	if e.cfg.ShowTiming {
		opts = append(opts, checkmate.WithShowTiming())
	}

	runner := checkmate.NewRunner(e.printer, opts...)

	for _, c := range checks {
		runner.AddFunc(c.name, c.fn).WithRemediation(c.remediation)
	}

	return runner.Run(ctx)
}

// finalResult returns an error if any checks failed.
func (e *Executor) finalResult(passed, failed, total int) error {
	if failed > 0 {
		return fmt.Errorf("%d/%d checks failed", failed, total)
	}
	return nil
}
