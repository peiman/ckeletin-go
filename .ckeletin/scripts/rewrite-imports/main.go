// AST-based Go import rewriter for ckeletin-go framework.
//
// Replaces module paths in import statements using go/ast, ensuring that only
// actual import paths are modified — never comments, string constants, or
// partial matches. Supports a -preserve-pkg flag for scaffold init, which
// keeps oldModule/pkg/* imports unchanged (those become external dependencies).
//
// Usage:
//
//	go run ./.ckeletin/scripts/rewrite-imports/ -old <module> -new <module> [-dir <dir>] [-preserve-pkg]
//
// Example:
//
//	go run ./.ckeletin/scripts/rewrite-imports/ \
//	    -old github.com/peiman/ckeletin-go \
//	    -new github.com/user/myapp \
//	    -dir .ckeletin
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	oldModule := flag.String("old", "", "Old module path to replace")
	newModule := flag.String("new", "", "New module path")
	dir := flag.String("dir", ".", "Directory to process")
	preservePkg := flag.Bool("preserve-pkg", false, "Preserve oldModule/pkg/ imports (for scaffold init)")
	flag.Parse()

	if *oldModule == "" || *newModule == "" {
		fmt.Fprintf(os.Stderr, "Usage: rewrite-imports -old <module> -new <module> [-dir <dir>] [-preserve-pkg]\n")
		os.Exit(1)
	}

	count, err := rewriteDir(*dir, *oldModule, *newModule, *preservePkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rewrote imports in %d files\n", count)
}

// rewriteDir walks a directory tree and rewrites imports in all .go files.
// It skips vendor, .git, dist, .task, and node_modules directories.
func rewriteDir(dir, oldModule, newModule string, preservePkg bool) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || base == ".git" || base == "dist" || base == ".task" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		changed, rewriteErr := rewriteFile(path, oldModule, newModule, preservePkg)
		if rewriteErr != nil {
			return fmt.Errorf("rewriting %s: %w", path, rewriteErr)
		}
		if changed {
			count++
		}
		return nil
	})
	return count, err
}

// rewriteFile parses a single Go file, rewrites matching import paths, and
// writes the result back. Returns true if any imports were changed.
func rewriteFile(path, oldModule, newModule string, preservePkg bool) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("parsing %s: %w", path, err)
	}

	changed := false
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Skip pkg/ imports if preserving (for scaffold init)
		if preservePkg && strings.HasPrefix(importPath, oldModule+"/pkg/") {
			continue
		}

		if importPath == oldModule || strings.HasPrefix(importPath, oldModule+"/") {
			newPath := newModule + importPath[len(oldModule):]
			imp.Path.Value = fmt.Sprintf(`"%s"`, newPath)
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	ast.SortImports(fset, node)

	f, err := os.Create(path) // #nosec G304 - path from filepath.Walk
	if err != nil {
		return false, fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	if err := format.Node(f, fset, node); err != nil {
		return false, fmt.Errorf("formatting %s: %w", path, err)
	}

	return true, nil
}
