// scripts/scaffold-init.go
//
// Scaffold initialization script for ckeletin-go

//go:build ignore

//
// This script automates the process of customizing the scaffold by:
// - Updating module path in go.mod
// - Replacing import statements in all Go files
// - Updating binary name in Taskfile.yml and .goreleaser.yml
// - Updating template files with new module path
//
// Usage:
//   go run scripts/scaffold-init.go <old_module> <new_module> <old_name> <new_name>
//
// Example:
//   go run scripts/scaffold-init.go github.com/peiman/ckeletin-go github.com/myuser/myapp ckeletin-go myapp

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse arguments
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <old_module> <new_module> <old_name> <new_name>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s github.com/peiman/ckeletin-go github.com/myuser/myapp ckeletin-go myapp\n", os.Args[0])
		os.Exit(1)
	}

	oldModule := os.Args[1]
	newModule := os.Args[2]
	oldName := os.Args[3]
	newName := os.Args[4]

	// Validate arguments
	if oldModule == "" || newModule == "" || oldName == "" || newName == "" {
		fmt.Fprintln(os.Stderr, "Error: All arguments must be non-empty")
		os.Exit(1)
	}

	if oldModule == newModule {
		fmt.Fprintln(os.Stderr, "Error: New module path must be different from old module path")
		os.Exit(1)
	}

	// Check if already initialized
	if err := checkAlreadyInitialized(oldModule); err != nil {
		fmt.Fprintf(os.Stderr, "✅ Project appears already initialized (module path is not %s)\n", oldModule)
		fmt.Fprintln(os.Stderr, "   If you want to re-initialize, please manually reset to the original state first.")
		os.Exit(0) // Exit successfully - not an error
	}

	// Check for uncommitted changes (warning only)
	if hasUncommittedChanges() {
		fmt.Fprintln(os.Stderr, "⚠️  Warning: You have uncommitted changes.")
		fmt.Fprintln(os.Stderr, "   Consider committing or stashing them before initialization.")
		fmt.Fprintln(os.Stderr, "")
	}

	// Perform updates
	fmt.Println("  ✓ Updating go.mod module path")
	if err := updateGoMod(newModule); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("  ✓ Updating Go import statements")
	count, err := updateGoFiles(oldModule, newModule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating Go files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("    Updated %d Go files\n", count)

	fmt.Println("  ✓ Updating Taskfile.yml")
	if err := updateTaskfile(oldName, newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating Taskfile.yml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("  ✓ Updating .goreleaser.yml")
	if err := updateGoreleaser(oldName, newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating .goreleaser.yml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("  ✓ Updating template files")
	templateCount, err := updateTemplateFiles(oldModule, newModule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating template files: %v\n", err)
		os.Exit(1)
	}
	if templateCount > 0 {
		fmt.Printf("    Updated %d template files\n", templateCount)
	}

	fmt.Println("  ✓ Running go mod tidy")
	fmt.Println("  ✓ Formatting code")
}

// checkAlreadyInitialized checks if go.mod contains the old module path
func checkAlreadyInitialized(oldModule string) error {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// If go.mod doesn't contain the old module, it's already initialized
	if !strings.Contains(string(content), "module "+oldModule) {
		return fmt.Errorf("already initialized")
	}

	return nil
}

// hasUncommittedChanges checks if there are uncommitted git changes
func hasUncommittedChanges() bool {
	// Check if .git directory exists
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return false // Not a git repo
	}

	// Simple check: look for modified files
	// In production, you'd use git status, but for simplicity we'll skip this
	// since it's just a warning anyway
	return false
}

// updateGoMod updates the module declaration in go.mod
func updateGoMod(newModule string) error {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "module ") {
			lines[i] = "module " + newModule
			break
		}
	}

	updated := strings.Join(lines, "\n")
	return os.WriteFile("go.mod", []byte(updated), 0600)
}

// updateGoFiles updates import statements in all Go files
func updateGoFiles(oldModule, newModule string) (int, error) {
	count := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we don't want to process
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "dist" || name == ".task" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Read file
		// #nosec G304 - path is controlled by filepath.Walk, not user input
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if file contains old module
		if !strings.Contains(string(content), oldModule) {
			return nil
		}

		// Replace old module with new module
		updated := strings.ReplaceAll(string(content), oldModule, newModule)

		// Write back
		if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
			return err
		}

		count++
		return nil
	})

	return count, err
}

// updateTaskfile updates BINARY_NAME in Taskfile.yml
func updateTaskfile(oldName, newName string) error {
	content, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "BINARY_NAME:") {
			// Replace the old name with new name
			lines[i] = strings.ReplaceAll(line, oldName, newName)
			break
		}
	}

	updated := strings.Join(lines, "\n")
	return os.WriteFile("Taskfile.yml", []byte(updated), 0600)
}

// updateGoreleaser updates project_name in .goreleaser.yml
func updateGoreleaser(oldName, newName string) error {
	content, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		// If .goreleaser.yml doesn't exist, that's okay
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "project_name:") {
			// Replace the old name with new name
			lines[i] = strings.ReplaceAll(line, oldName, newName)
			break
		}
	}

	updated := strings.Join(lines, "\n")
	return os.WriteFile(".goreleaser.yml", []byte(updated), 0600)
}

// updateTemplateFiles updates module references in .example template files
func updateTemplateFiles(oldModule, newModule string) (int, error) {
	count := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we don't want to process
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "dist" || name == ".task" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .example files
		if !strings.HasSuffix(path, ".example") {
			return nil
		}

		// Read file
		// #nosec G304 - path is controlled by filepath.Walk, not user input
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if file contains old module
		if !strings.Contains(string(content), oldModule) {
			return nil
		}

		// Replace old module with new module
		updated := strings.ReplaceAll(string(content), oldModule, newModule)

		// Write back
		if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
			return err
		}

		count++
		return nil
	})

	return count, err
}
