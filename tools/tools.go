//go:build tools
// +build tools

// Package tools is used to track binary dependencies with go modules.
package tools

// Import packages that we want `go mod` to download for us.
// These are build-time dependencies, NOT runtime dependencies!
import (
	// Tools we use
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "gotest.tools/gotestsum"
)

// Note: Some tools are installed directly via the Makefile as they're binaries:
// - golangci-lint
// - richgo
