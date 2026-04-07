// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMainFunction(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		cmd      string
		cmdError error
		wantCode int
	}{
		{
			name:     "Success scenario",
			cmd:      "success",
			cmdError: nil,
			wantCode: 0,
		},
		{
			name:     "Failure scenario",
			cmd:      "fail",
			cmdError: fmt.Errorf("simulated failure"),
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save the original RootCmd
			originalRoot := cmd.RootCmd
			// Create a test root command
			testRoot := &cobra.Command{Use: "test"}
			// Replace global RootCmd with our test root
			cmd.RootCmd = testRoot
			// Restore after the test
			defer func() { cmd.RootCmd = originalRoot }()

			// Add a dummy command with the specified behavior
			testRoot.AddCommand(&cobra.Command{
				Use: tt.cmd,
				RunE: func(cmd *cobra.Command, args []string) error {
					return tt.cmdError
				},
			})

			// Set command arguments
			testRoot.SetArgs([]string{tt.cmd})

			// EXECUTION PHASE
			code := run()

			// ASSERTION PHASE
			if code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, code)
			}
		})
	}
}

func TestRun_JSONMode_Error(t *testing.T) {
	originalRoot := cmd.RootCmd
	defer func() { cmd.RootCmd = originalRoot }()

	output.SetOutputMode("json")
	output.SetCommandName("fail")
	defer func() {
		output.SetOutputMode("")
		output.SetCommandName("")
	}()

	testRoot := &cobra.Command{Use: "test", SilenceErrors: true}
	testRoot.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("simulated failure")
		},
	})
	testRoot.SetArgs([]string{"fail"})
	cmd.RootCmd = testRoot

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := run()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Equal(t, 1, code)

	var envelope output.JSONEnvelope
	err := json.Unmarshal(buf.Bytes(), &envelope)
	assert.NoError(t, err, "stdout should contain valid JSON, got: %s", buf.String())
	assert.Equal(t, "error", envelope.Status)
	assert.NotNil(t, envelope.Error)
	assert.Contains(t, envelope.Error.Message, "simulated failure")
}
