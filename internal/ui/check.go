package ui

import (
	"io"
	"os"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

// NewCheckPrinter creates a new checkmate.Printer configured for this project.
// Use this for check-style output (validations, quality checks, build steps).
//
// Example:
//
//	p := ui.NewCheckPrinter()
//	p.CategoryHeader("Code Quality")
//	p.CheckHeader("Running linter")
//	p.CheckSuccess("No issues found")
func NewCheckPrinter(opts ...checkmate.Option) *checkmate.Printer {
	return checkmate.New(opts...)
}

// NewCheckPrinterWithWriter creates a checkmate.Printer that writes to the specified writer.
func NewCheckPrinterWithWriter(w io.Writer, opts ...checkmate.Option) *checkmate.Printer {
	opts = append([]checkmate.Option{checkmate.WithWriter(w)}, opts...)
	return checkmate.New(opts...)
}

// StdoutCheckPrinter returns a checkmate.Printer that writes to stdout.
// This is a convenience for the common case.
func StdoutCheckPrinter() *checkmate.Printer {
	return checkmate.New(checkmate.WithWriter(os.Stdout))
}

// StderrCheckPrinter returns a checkmate.Printer that writes to stderr.
func StderrCheckPrinter() *checkmate.Printer {
	return checkmate.New(checkmate.WithStderr())
}
