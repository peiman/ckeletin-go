package main

import (
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

// parseModuleParts extracts the second and third path segments from a Go module path.
// For the standard "host/owner/repo" pattern, these correspond to owner and repo.
// For "github.com/owner/repo", returns ("owner", "repo").
// For "github.com/org/repo/v2", returns ("org", "repo") — segments beyond the third are ignored.
// For "example.com/tool", returns ("", "tool").
// For "mymodule", returns ("", "mymodule").
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

// cleanArchLintConfig removes the public component from .go-arch-lint.yml.
// Called after pkg/ is removed during scaffold init, since the public component
// references pkg/** which no longer exists.
func cleanArchLintConfig(projectRoot string) error {
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
