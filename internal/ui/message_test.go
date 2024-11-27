// internal/ui/message_test.go

package ui

import (
	"bytes"
	"testing"
)

func TestPrintColoredMessage(t *testing.T) {
	buf := new(bytes.Buffer)
	err := PrintColoredMessage(buf, "Test Message", "green")
	if err != nil {
		t.Fatalf("PrintColoredMessage returned an error: %v", err)
	}

	output := buf.String()
	expected := "Test Message"
	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected output to contain %q, got %q", expected, output)
	}
}
