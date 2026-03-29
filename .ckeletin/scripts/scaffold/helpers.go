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
