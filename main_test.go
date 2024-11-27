package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/peiman/ckeletin-go/cmd"
)

func TestRun(t *testing.T) {
	// Create a pipe to capture stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect stdout
	stdout := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = stdout }()

	// Run the command
	cmd.RootCommand().SetArgs([]string{"--help"}) // Example: Test the help command
	run()

	// Close the writer and read the captured output
	writer.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	// Validate output
	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected output, but got none")
	}

	// Optionally, check for specific content in the output
	if expected := "Usage:"; !bytes.Contains(buf.Bytes(), []byte(expected)) {
		t.Errorf("Expected output to contain %q, but it didn't", expected)
	}
}
