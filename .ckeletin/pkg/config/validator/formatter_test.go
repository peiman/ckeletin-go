// internal/config/validator/formatter_test.go

package validator

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatResult(t *testing.T) {
	tests := []struct {
		name            string
		result          *Result
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "Valid config with no warnings",
			result: &Result{
				Valid:      true,
				Errors:     []error{},
				Warnings:   []string{},
				ConfigFile: "/path/to/config.yaml",
			},
			wantContains: []string{
				"Validating: /path/to/config.yaml",
				"✅ Configuration is valid!",
			},
			wantNotContains: []string{
				"❌ Errors",
				"⚠️  Warnings",
			},
		},
		{
			name: "Valid config with warnings",
			result: &Result{
				Valid:      true,
				Errors:     []error{},
				Warnings:   []string{"Unknown key: foo", "Deprecated option: bar"},
				ConfigFile: "/path/to/config.yaml",
			},
			wantContains: []string{
				"Validating: /path/to/config.yaml",
				"⚠️  Warnings (2)",
				"1. Unknown key: foo",
				"2. Deprecated option: bar",
				"✅ Configuration is valid (with warnings)",
			},
			wantNotContains: []string{
				"❌ Errors",
			},
		},
		{
			name: "Invalid config with errors",
			result: &Result{
				Valid:      false,
				Errors:     []error{errors.New("invalid syntax"), errors.New("missing required field")},
				Warnings:   []string{},
				ConfigFile: "/path/to/config.yaml",
			},
			wantContains: []string{
				"Validating: /path/to/config.yaml",
				"❌ Errors (2)",
				"1. invalid syntax",
				"2. missing required field",
				"❌ Configuration is invalid",
			},
			wantNotContains: []string{
				"⚠️  Warnings",
			},
		},
		{
			name: "Invalid config with errors and warnings",
			result: &Result{
				Valid:      false,
				Errors:     []error{errors.New("parse error")},
				Warnings:   []string{"Unknown key: test"},
				ConfigFile: "/path/to/config.yaml",
			},
			wantContains: []string{
				"Validating: /path/to/config.yaml",
				"❌ Errors (1)",
				"1. parse error",
				"⚠️  Warnings (1)",
				"1. Unknown key: test",
				"❌ Configuration is invalid",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			FormatResult(tt.result, &buf)
			output := buf.String()

			// Check for expected content
			for _, want := range tt.wantContains {
				assert.True(t, strings.Contains(output, want),
					"FormatResult() output missing expected content %q\nGot: %s", want, output)
			}

			// Check for unexpected content
			for _, notWant := range tt.wantNotContains {
				assert.False(t, strings.Contains(output, notWant),
					"FormatResult() output contains unexpected content %q\nGot: %s", notWant, output)
			}
		})
	}
}

func TestExitCodeForResult(t *testing.T) {
	tests := []struct {
		name    string
		result  *Result
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid with no warnings - exit 0",
			result: &Result{
				Valid:    true,
				Errors:   []error{},
				Warnings: []string{},
			},
			wantErr: false,
		},
		{
			name: "Valid with warnings - exit 1",
			result: &Result{
				Valid:    true,
				Errors:   []error{},
				Warnings: []string{"warning1"},
			},
			wantErr: true,
			errMsg:  "validation completed with warnings",
		},
		{
			name: "Invalid with errors - exit 1",
			result: &Result{
				Valid:    false,
				Errors:   []error{errors.New("error1")},
				Warnings: []string{},
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "Invalid with errors and warnings - exit 1",
			result: &Result{
				Valid:    false,
				Errors:   []error{errors.New("error1")},
				Warnings: []string{"warning1"},
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExitCodeForResult(tt.result)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
