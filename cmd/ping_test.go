// cmd/ping_test.go
package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/internal/ui"
)

func TestPrintColoredMessage(t *testing.T) {
	err := printColoredMessage("Test Message", "green")
	if err != nil {
		t.Errorf("Expected no error for valid color, got %v", err)
	}

	err = printColoredMessage("Test Message", "invalid-color")
	if err == nil {
		t.Errorf("Expected error for invalid color, got nil")
	}
}

func TestGetLipglossColor(t *testing.T) {
	_, err := ui.GetLipglossColor("green")
	if err != nil {
		t.Errorf("Expected no error for valid color, got %v", err)
	}

	_, err = ui.GetLipglossColor("invalid-color")
	if err == nil {
		t.Errorf("Expected error for invalid color, got nil")
	}
}
