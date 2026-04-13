package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// replaceModulePreservingPkg replaces oldModule with newModule in content,
// but preserves lines that reference oldModule/pkg/ (external library imports).
// This allows derived projects to keep pkg/ packages as external dependencies
// from the original ckeletin-go module.
func replaceModulePreservingPkg(content, oldModule, newModule string) string {
	pkgPrefix := oldModule + "/pkg/"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pkgPrefix) {
			continue // Preserve pkg/ references
		}
		lines[i] = strings.ReplaceAll(line, oldModule, newModule)
	}
	return strings.Join(lines, "\n")
}

// removePkgDirectory removes the pkg/ directory from the project root.
// After scaffold init, pkg/ packages (like checkmate) are consumed as external
// dependencies from the original ckeletin-go module, not local copies.
func removePkgDirectory(projectRoot string) error {
	pkgDir := filepath.Join(projectRoot, "pkg")
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return nil // Nothing to remove
	}
	return os.RemoveAll(pkgDir)
}

// removeFrameworkOnlyArtifacts removes files and directories that are specific
// to the ckeletin-go framework development and should not be in downstream projects.
// This includes conformance testing (spec compliance), scaffold integration tests,
// and the conformance mapping.
func removeFrameworkOnlyArtifacts(projectRoot string) error {
	artifacts := []string{
		"test/conformance",
		"conformance-mapping.yaml",
	}
	for _, artifact := range artifacts {
		path := filepath.Join(projectRoot, artifact)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("removing %s: %w", artifact, err)
		}
	}

	// Remove scaffold build tag from integration tests — downstream projects
	// keep the integration tests but don't need the scaffold tag since they
	// won't run scaffold init on themselves.
	scaffoldTest := filepath.Join(projectRoot, "test", "integration", "scaffold_init_test.go")
	if _, err := os.Stat(scaffoldTest); err == nil {
		if err := os.Remove(scaffoldTest); err != nil {
			return fmt.Errorf("removing scaffold_init_test.go: %w", err)
		}
	}

	return nil
}

// parseModuleParts extracts the second and third path segments from a Go module path.
// For the standard "host/owner/repo" pattern, these correspond to owner and repo.
// For "github.com/owner/repo", returns ("owner", "repo").
// For "github.com/org/repo/v2", returns ("org", "repo") — segments beyond the third are ignored.
// For "example.com/tool", returns ("", "tool").
// For "mymodule", returns ("", "mymodule").
// toEnvPrefix converts a binary name to its environment variable prefix form.
// Uppercases the name and replaces non-alphanumeric characters with underscores.
// For "ckeletin-go" returns "CKELETIN_GO", for "myapp" returns "MYAPP".
func toEnvPrefix(name string) string {
	upper := strings.ToUpper(name)
	var result strings.Builder
	for _, r := range upper {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}

func parseModuleParts(module string) (owner, repo string) {
	parts := strings.Split(module, "/")
	switch {
	case len(parts) >= 3:
		return parts[1], parts[2]
	case len(parts) == 2:
		return "", parts[1]
	default:
		return "", module
	}
}

// resetChangelog replaces CHANGELOG.md with an empty keepachangelog template.
func resetChangelog(projectRoot string) error {
	const changelogTemplate = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
`
	return os.WriteFile(filepath.Join(projectRoot, "CHANGELOG.md"), []byte(changelogTemplate), 0600)
}

// resetLicense replaces LICENSE with an MIT template containing placeholder values.
func resetLicense(projectRoot string) error {
	const licenseTemplate = `MIT License

Copyright (c) [YEAR] [YOUR NAME OR COMPANY]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
	return os.WriteFile(filepath.Join(projectRoot, "LICENSE"), []byte(licenseTemplate), 0600)
}

// StringReplacement defines a single find/replace pair.
type StringReplacement struct {
	Old string
	New string
}

// replaceInTextFiles walks the project tree and applies string replacements
// sequentially in text files (.md, .yml, .yaml). It skips .git, vendor,
// dist, and .task directories. Replacements are applied in slice order,
// so callers should list most-specific patterns first to prevent shorter
// patterns from corrupting longer ones' match targets.
// Returns the number of files that were modified.
func replaceInTextFiles(root string, replacements []StringReplacement) (int, error) {
	count := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "dist" || name == ".task" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process text files we care about
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".yml" && ext != ".yaml" {
			return nil
		}

		// #nosec G304 - path is controlled by filepath.Walk, not user input
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		updated := original

		for _, r := range replacements {
			updated = strings.ReplaceAll(updated, r.Old, r.New)
		}

		if updated == original {
			return nil
		}

		if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
			return err
		}

		count++
		return nil
	})

	return count, err
}

// replaceNameInGoFiles walks the project tree and replaces the old binary name
// with the new name in Go source files, skipping import statements.
// Import paths are handled separately by the AST-based import rewriter
// (which deliberately preserves pkg/ imports), so this function must not
// modify import lines. It handles string literals, comments, and other
// non-import references.
// Returns the number of files that were modified.
func replaceNameInGoFiles(root, oldName, newName string) (int, error) {
	count := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "dist" || name == ".task" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// #nosec G304 - path is controlled by filepath.Walk, not user input
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)

		// Replace line by line, skipping import lines to avoid clobbering
		// preserved pkg/ imports from the AST rewriter
		lines := strings.Split(original, "\n")
		inImportBlock := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Track import blocks
			if strings.HasPrefix(trimmed, "import (") {
				inImportBlock = true
				continue
			}
			if inImportBlock {
				if trimmed == ")" {
					inImportBlock = false
				}
				continue
			}
			// Skip single-line imports
			if strings.HasPrefix(trimmed, "import ") {
				continue
			}

			lines[i] = strings.ReplaceAll(line, oldName, newName)
		}

		updated := strings.Join(lines, "\n")

		if updated == original {
			return nil
		}

		if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
			return err
		}

		count++
		return nil
	})

	return count, err
}

// cleanArchLintConfig removes the public component from .go-arch-lint.yml
// and registers the upstream module's pkg/ as a vendor dependency.
// Called after pkg/ is removed during scaffold init, since the public component
// references pkg/** which no longer exists. The upstream pkg/ is now consumed
// as an external dependency via replace directive.
func cleanArchLintConfig(projectRoot, oldModule string) error {
	configPath := filepath.Join(projectRoot, ".go-arch-lint.yml")

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	original := string(content)
	lines := strings.Split(original, "\n")
	var result []string
	inPublicBlock := false
	publicIndent := 0

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Skip comment lines related to the public section
		stripped := strings.TrimSpace(trimmed)
		if strings.Contains(stripped, "PUBLIC PACKAGES") ||
			strings.Contains(stripped, "pkg/ contains standalone") ||
			strings.Contains(stripped, "Do NOT import from internal/") ||
			strings.Contains(stripped, "Can be imported by external") ||
			strings.Contains(stripped, "Are independent of the CLI") ||
			strings.Contains(stripped, "See ADR-010 for guidance") ||
			strings.Contains(stripped, "Cannot depend on any internal packages (enforced by validation script too)") ||
			strings.Contains(stripped, "Can use any external vendor dependencies") ||
			strings.Contains(stripped, "Public packages are completely standalone") {
			continue
		}

		// Skip separator lines that are part of the public block header
		if strings.HasPrefix(stripped, "# -----") && !inPublicBlock {
			// Look ahead: if next meaningful content is public-related, skip
			// For now, keep it — the comment lines above catch the specific ones
		}

		// Detect "  public:" blocks (under components: or deps:)
		if trimmed == "  public:" {
			inPublicBlock = true
			publicIndent = 2
			continue
		}

		if inPublicBlock {
			if trimmed == "" {
				continue
			}
			strippedLine := strings.TrimLeft(line, " ")
			indent := len(line) - len(strippedLine)
			if indent > publicIndent {
				continue // Still inside the block
			}
			inPublicBlock = false
		}

		// Skip "- public" entries in commonComponents
		if strings.TrimSpace(trimmed) == "- public" ||
			strings.Contains(trimmed, "- public  #") ||
			strings.Contains(trimmed, "- public\t#") {
			continue
		}

		result = append(result, line)
	}

	updated := strings.Join(result, "\n")

	if updated == original {
		return nil
	}

	return os.WriteFile(configPath, []byte(updated), 0600)
}

// registerUpstreamVendor adds the upstream module's pkg/ as a vendor dependency
// in .go-arch-lint.yml. Called AFTER text replacement so the upstream module
// path isn't rewritten to the new module path.
func registerUpstreamVendor(projectRoot, oldModule string) error {
	configPath := filepath.Join(projectRoot, ".go-arch-lint.yml")

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	updated := string(content)

	// Check if depOnAnyVendor is false (vendor registration only needed then)
	if !strings.Contains(updated, "depOnAnyVendor: false") {
		return nil
	}

	pkgVendor := oldModule + "/pkg/**"

	// Add vendor entry before commonComponents
	vendorBlock := fmt.Sprintf(
		"\n  # Upstream framework public packages\n  ckeletin-pkg:\n    in: %s\n",
		pkgVendor)

	if idx := strings.Index(updated, "commonComponents:"); idx != -1 {
		insertAt := strings.LastIndex(updated[:idx], "\n")
		if insertAt != -1 {
			updated = updated[:insertAt] + vendorBlock + updated[insertAt:]
		}
	}

	// Add ckeletin-pkg to canUse for business and infrastructure
	lines := strings.Split(updated, "\n")
	var finalLines []string
	inDepsSection := false
	currentComponent := ""

	for i, line := range lines {
		finalLines = append(finalLines, line)
		trimmed := strings.TrimSpace(line)

		if trimmed == "deps:" {
			inDepsSection = true
		}

		if inDepsSection {
			// Component names are at 2-space indent (direct children of deps:)
			if strings.HasSuffix(line, ":") && len(line) > 2 && line[:2] == "  " && line[2] != ' ' {
				currentComponent = strings.TrimSuffix(strings.TrimSpace(line), ":")
			}

			if trimmed == "canUse:" && (currentComponent == "business" || currentComponent == "infrastructure") {
				if i+1 < len(lines) && !strings.Contains(lines[i+1], "ckeletin-pkg") {
					finalLines = append(finalLines, "      - ckeletin-pkg")
				}
			}
		}
	}

	updated = strings.Join(finalLines, "\n")

	if updated == string(content) {
		return nil
	}

	return os.WriteFile(configPath, []byte(updated), 0600)
}
