// internal/dev/doctor.go
//
// Environment health checker for development builds.
// Verifies development environment is properly configured.

package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name    string
	Status  CheckStatus
	Message string
	Details string
}

// CheckStatus represents the status of a health check
type CheckStatus int

const (
	CheckPassed CheckStatus = iota
	CheckWarning
	CheckFailed
)

// Styling for check statuses
var (
	passStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))   // Yellow
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // Red
	detailStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

// String returns a colorized string representation of CheckStatus
func (s CheckStatus) String() string {
	switch s {
	case CheckPassed:
		return passStyle.Render("✓ PASS")
	case CheckWarning:
		return warnStyle.Render("⚠ WARN")
	case CheckFailed:
		return failStyle.Render("✗ FAIL")
	default:
		return "?"
	}
}

// Doctor performs environment health checks
type Doctor struct {
	checks []HealthCheck
}

// NewDoctor creates a new Doctor instance
func NewDoctor() *Doctor {
	return &Doctor{
		checks: []HealthCheck{},
	}
}

// RunAllChecks executes all health checks
func (d *Doctor) RunAllChecks() {
	d.checks = []HealthCheck{}

	// Check required tools
	d.checkTool("task", "Task runner")
	d.checkTool("go", "Go compiler")
	d.checkTool("golangci-lint", "Go linter")
	d.checkTool("gotestsum", "Test runner")
	d.checkTool("go-licenses", "License checker (source)")
	d.checkTool("lichen", "License checker (binary)")
	d.checkTool("git", "Version control")

	// Check Go version
	d.checkGoVersion()

	// Check project structure
	d.checkProjectStructure()

	// Check git status
	d.checkGitStatus()

	// Check dependencies
	d.checkDependencies()
}

// GetResults returns all health check results
func (d *Doctor) GetResults() []HealthCheck {
	return d.checks
}

// HasFailures returns true if any check failed
func (d *Doctor) HasFailures() bool {
	for _, check := range d.checks {
		if check.Status == CheckFailed {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any check has warnings
func (d *Doctor) HasWarnings() bool {
	for _, check := range d.checks {
		if check.Status == CheckWarning {
			return true
		}
	}
	return false
}

// FormatResults returns a formatted string of all check results
func (d *Doctor) FormatResults() string {
	var sb strings.Builder

	sb.WriteString("Development Environment Health Check\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n\n")

	for _, check := range d.checks {
		sb.WriteString(fmt.Sprintf("%s %s\n", check.Status.String(), check.Name))
		if check.Message != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", check.Message))
		}
		if check.Details != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", detailStyle.Render(check.Details)))
		}
		sb.WriteString("\n")
	}

	// Summary
	passed := 0
	warnings := 0
	failed := 0
	for _, check := range d.checks {
		switch check.Status {
		case CheckPassed:
			passed++
		case CheckWarning:
			warnings++
		case CheckFailed:
			failed++
		}
	}

	sb.WriteString(strings.Repeat("-", 80) + "\n")
	sb.WriteString(fmt.Sprintf("Summary: %d passed, %d warnings, %d failed\n",
		passed, warnings, failed))

	if failed > 0 {
		sb.WriteString("\n" + errorStyle.Render("❌ Environment has issues that need attention") + "\n")
	} else if warnings > 0 {
		sb.WriteString("\n" + warnStyle.Render("⚠️  Environment is functional but has warnings") + "\n")
	} else {
		sb.WriteString("\n" + successStyle.Render("✅ Environment is healthy") + "\n")
	}

	return sb.String()
}

// checkTool verifies a tool is installed and accessible
func (d *Doctor) checkTool(name, description string) {
	path, err := exec.LookPath(name)
	if err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    description,
			Status:  CheckFailed,
			Message: fmt.Sprintf("%s not found in PATH", name),
			Details: "Run 'task setup' to install development tools",
		})
		return
	}

	// Get version if possible
	version := d.getToolVersion(name)
	d.checks = append(d.checks, HealthCheck{
		Name:    description,
		Status:  CheckPassed,
		Message: fmt.Sprintf("%s found", name),
		Details: fmt.Sprintf("Path: %s | %s", path, version),
	})
}

// getToolVersion attempts to get version information for a tool
func (d *Doctor) getToolVersion(name string) string {
	var cmd *exec.Cmd

	switch name {
	case "go":
		cmd = exec.Command(name, "version")
	case "task":
		cmd = exec.Command(name, "--version")
	case "golangci-lint":
		cmd = exec.Command(name, "version")
	case "gotestsum":
		cmd = exec.Command(name, "--version")
	case "go-licenses":
		cmd = exec.Command(name, "--version")
	case "lichen":
		cmd = exec.Command(name, "--version")
	case "git":
		cmd = exec.Command(name, "--version")
	default:
		return ""
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "version unknown"
	}

	version := strings.TrimSpace(string(output))
	// Truncate if too long
	if len(version) > 60 {
		version = version[:57] + "..."
	}
	return version
}

// checkGoVersion verifies Go version matches requirements
func (d *Doctor) checkGoVersion() {
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Go version",
			Status:  CheckFailed,
			Message: "Failed to get Go version",
			Details: err.Error(),
		})
		return
	}

	versionStr := string(output)
	// Extract version number (e.g., "go1.25.3" from "go version go1.25.3 darwin/arm64")
	re := regexp.MustCompile(`go(\d+\.\d+)`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 2 {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Go version",
			Status:  CheckWarning,
			Message: "Could not parse Go version",
			Details: versionStr,
		})
		return
	}

	version := matches[1]
	// Check if version is 1.25+
	if version < "1.25" {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Go version",
			Status:  CheckWarning,
			Message: fmt.Sprintf("Go %s found, but 1.25+ recommended", version),
			Details: "Project requires Go 1.25 or higher",
		})
		return
	}

	d.checks = append(d.checks, HealthCheck{
		Name:    "Go version",
		Status:  CheckPassed,
		Message: fmt.Sprintf("Go %s meets requirements", version),
		Details: strings.TrimSpace(versionStr),
	})
}

// checkProjectStructure verifies project structure is valid
func (d *Doctor) checkProjectStructure() {
	requiredDirs := []string{"cmd", "internal", "docs/adr", "test/integration"}
	requiredFiles := []string{"go.mod", "Taskfile.yml", "README.md", "CLAUDE.md"}

	missingDirs := []string{}
	missingFiles := []string{}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			missingDirs = append(missingDirs, dir)
		}
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingDirs) > 0 || len(missingFiles) > 0 {
		details := []string{}
		if len(missingDirs) > 0 {
			details = append(details, fmt.Sprintf("Missing dirs: %s", strings.Join(missingDirs, ", ")))
		}
		if len(missingFiles) > 0 {
			details = append(details, fmt.Sprintf("Missing files: %s", strings.Join(missingFiles, ", ")))
		}

		d.checks = append(d.checks, HealthCheck{
			Name:    "Project structure",
			Status:  CheckFailed,
			Message: "Required directories or files missing",
			Details: strings.Join(details, " | "),
		})
		return
	}

	// Count ADRs
	adrFiles, err := filepath.Glob("docs/adr/*.md")
	adrCount := 0
	if err == nil {
		adrCount = len(adrFiles)
	}

	d.checks = append(d.checks, HealthCheck{
		Name:    "Project structure",
		Status:  CheckPassed,
		Message: "All required directories and files present",
		Details: fmt.Sprintf("%d ADRs found", adrCount),
	})
}

// checkGitStatus verifies git repository status
func (d *Doctor) checkGitStatus() {
	// Check if in git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Git repository",
			Status:  CheckFailed,
			Message: "Not a git repository",
			Details: "Initialize with 'git init'",
		})
		return
	}

	// Get current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Git repository",
			Status:  CheckWarning,
			Message: "Could not determine current branch",
			Details: err.Error(),
		})
		return
	}

	branch := strings.TrimSpace(string(output))

	// Get dirty status
	cmd = exec.Command("git", "status", "--porcelain")
	output, err = cmd.CombinedOutput()
	if err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Git repository",
			Status:  CheckWarning,
			Message: "Could not check git status",
			Details: err.Error(),
		})
		return
	}

	isDirty := len(strings.TrimSpace(string(output))) > 0
	dirtyStr := "clean"
	if isDirty {
		dirtyStr = "uncommitted changes"
	}

	d.checks = append(d.checks, HealthCheck{
		Name:    "Git repository",
		Status:  CheckPassed,
		Message: "Git repository initialized",
		Details: fmt.Sprintf("Branch: %s | Status: %s", branch, dirtyStr),
	})
}

// checkDependencies verifies go.mod and dependencies are in sync
func (d *Doctor) checkDependencies() {
	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Dependencies",
			Status:  CheckFailed,
			Message: "go.mod not found",
			Details: "Initialize with 'go mod init'",
		})
		return
	}

	// Run go mod verify
	cmd := exec.Command("go", "mod", "verify")
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Dependencies",
			Status:  CheckFailed,
			Message: "Dependency verification failed",
			Details: string(output),
		})
		return
	}

	// Check if go.mod and go.sum are in sync (go mod tidy -check)
	cmd = exec.Command("go", "mod", "tidy", "-diff")
	output, _ = cmd.CombinedOutput()

	// If there's output, go.mod needs tidying
	// Note: go mod tidy -diff returns non-zero exit if changes needed, which is expected
	if len(strings.TrimSpace(string(output))) > 0 {
		d.checks = append(d.checks, HealthCheck{
			Name:    "Dependencies",
			Status:  CheckWarning,
			Message: "go.mod may need tidying",
			Details: "Run 'go mod tidy' to clean up",
		})
		return
	}

	d.checks = append(d.checks, HealthCheck{
		Name:    "Dependencies",
		Status:  CheckPassed,
		Message: "All dependencies verified and in sync",
		Details: strings.TrimSpace(string(output)),
	})
}
