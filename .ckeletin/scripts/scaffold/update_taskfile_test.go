package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// updateTaskfile rewrites BINARY_NAME in the project Taskfile.yml during `task init`.
// It is the scaffold-side of the binary-name -> config-dir glue: if it stops
// matching the BINARY_NAME line, a scaffolded project keeps BINARY_NAME=ckeletin-go
// and (via the ldflags chain) reads ~/.config/ckeletin-go. Until now it was only
// exercised by the gated //go:build scaffold integration test (which does not run
// in `task check`), so this pins the rename directly and cheaply.
func TestUpdateTaskfileRenamesBinaryName(t *testing.T) {
	// Mirror the real Taskfile.yml header: a descriptive comment that does NOT
	// itself contain "BINARY_NAME:", followed by the real var. updateTaskfile
	// rewrites the first line containing "BINARY_NAME:".
	const taskfile = `# Taskfile.yml - Project tasks
version: '3'

vars:
  # Project configuration - customize these for your CLI
  BINARY_NAME: ckeletin-go
  MODULE_PATH:
    sh: go list -m
`
	dir := t.TempDir()
	t.Chdir(dir) // updateTaskfile reads/writes "Taskfile.yml" relative to cwd
	require.NoError(t, os.WriteFile("Taskfile.yml", []byte(taskfile), 0o600))

	require.NoError(t, updateTaskfile("ckeletin-go", "myapp"))

	got, err := os.ReadFile("Taskfile.yml")
	require.NoError(t, err)
	out := string(got)

	assert.Contains(t, out, "BINARY_NAME: myapp",
		"BINARY_NAME must be renamed to the new project name")
	assert.NotContains(t, out, "BINARY_NAME: ckeletin-go",
		"the old binary name must be gone after rename")
	// The descriptive comment above the var must be left untouched (it does not
	// carry the old name, so a blanket replace would be a bug).
	assert.Contains(t, out, "# Project configuration - customize these for your CLI",
		"surrounding content must be preserved")
}
