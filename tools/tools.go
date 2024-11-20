//nolint:all // Ignore linting for this file
//go:build tools
// +build tools

// Package tools tracks binary dependencies using go modules.
// These imports are not used in the actual application code.
package tools

import (
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "gotest.tools/gotestsum"
)
