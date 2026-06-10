// internal/docs/json.go
//
// Machine-readable output mode for documentation generation (CKSPEC-OUT-002).
// The generated documentation is wrapped as the data payload of the standard
// success envelope (CKSPEC-OUT-003) instead of being streamed as raw text.

package docs

import (
	"bytes"
	"io"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/rs/zerolog/log"
)

// Result is the data payload of the JSON envelope emitted by GenerateJSON.
type Result struct {
	// Format is the documentation format that was generated (markdown or yaml).
	Format string `json:"format"`
	// OutputFile is set when the documentation was written to a file.
	OutputFile string `json:"output_file,omitempty"`
	// Content carries the generated documentation when no output file is set.
	Content string `json:"content,omitempty"`
}

// GenerateJSON produces the documentation and emits it to w as the data of a
// standard success envelope (CKSPEC-OUT-002). When an output file is
// configured, the documentation is written to that file as in text mode and
// the envelope reports the file path instead of inlining the content.
//
// On error nothing is written to w: the caller (main.go) renders the single
// error envelope, preserving the one-envelope-per-command contract.
func (g *Generator) GenerateJSON(w io.Writer) error {
	var buf bytes.Buffer
	savedWriter := g.cfg.Writer
	g.cfg.Writer = &buf
	defer func() { g.cfg.Writer = savedWriter }()
	if err := g.Generate(); err != nil {
		return err
	}

	result := Result{
		Format:     strings.ToLower(g.cfg.OutputFormat),
		OutputFile: g.cfg.OutputFile,
	}
	if g.cfg.OutputFile == "" {
		result.Content = buf.String()
	}

	// Shadow log (CKSPEC-OUT-004): record the rendered operation in the audit
	// stream. The artifact itself is reproducible from the registry, so the
	// audit entry records what was generated and where, not the full text.
	log.Debug().
		Str("component", "docs").
		Str("format", result.Format).
		Str("output_file", result.OutputFile).
		Int("content_bytes", len(result.Content)).
		Msg("Documentation generated in JSON output mode")

	return output.RenderJSON(w, output.JSONEnvelope{
		Status:  "success",
		Command: output.CommandName(),
		Data:    result,
	})
}
