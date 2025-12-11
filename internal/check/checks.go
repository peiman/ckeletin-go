// internal/check/checks.go
//
// Individual check implementations for the check command.

package check

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// checkFormat checks code formatting using goimports and gofmt
func (e *Executor) checkFormat(ctx context.Context) error {
	log.Debug().Msg("Running format check")

	// Check goimports
	cmd := exec.CommandContext(ctx, "goimports", "-l", ".")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("files need formatting:\n%s", strings.TrimSpace(string(output)))
	}

	// Check gofmt
	cmd = exec.CommandContext(ctx, "gofmt", "-l", ".")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("files need formatting:\n%s", strings.TrimSpace(string(output)))
	}

	return nil
}

// checkLint runs go vet and golangci-lint
func (e *Executor) checkLint(ctx context.Context) error {
	log.Debug().Msg("Running lint check")

	// go vet
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go vet failed:\n%s", strings.TrimSpace(string(output)))
	}

	// golangci-lint
	cmd = exec.CommandContext(ctx, "golangci-lint", "run")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("golangci-lint failed:\n%s", strings.TrimSpace(string(output)))
	}

	return nil
}

// checkTest runs tests with race detection
func (e *Executor) checkTest(ctx context.Context) error {
	log.Debug().Msg("Running test check")

	cmd := exec.CommandContext(ctx, "go", "test", "-race", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tests failed:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// checkDeps verifies dependency integrity
func (e *Executor) checkDeps(ctx context.Context) error {
	log.Debug().Msg("Running deps check")

	cmd := exec.CommandContext(ctx, "go", "mod", "verify")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dependency verification failed:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// checkVuln scans for vulnerabilities using govulncheck
func (e *Executor) checkVuln(ctx context.Context) error {
	log.Debug().Msg("Running vulnerability check")

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vulnerabilities found:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}
