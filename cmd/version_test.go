package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	assert.Contains(t, output, "ckeletin-go v")
	assert.Contains(t, output, Version)
	assert.Contains(t, output, Date)
	assert.Contains(t, output, Commit)
}
