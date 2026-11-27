package logger

import (
	"bytes"
	"encoding/json"
	"io"
)

// singleKeyJSON rewrites one JSON object so each key appears once. Zerolog
// writes a context field and then the same key on the event, which is two
// keys in the file line. The later value is the one that wins.
type singleKeyJSON struct {
	Writer io.Writer
}

func (w singleKeyJSON) Write(p []byte) (int, error) {
	collapsed := collapseJSONKeys(p)
	n, err := w.Writer.Write(collapsed)
	if err != nil {
		return n, err
	}
	return len(p), nil
}

func collapseJSONKeys(p []byte) []byte {
	trimmed := bytes.TrimSpace(p)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return p
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return p
	}
	out, err := json.Marshal(fields)
	if err != nil {
		return p
	}
	if len(p) > 0 && p[len(p)-1] == '\n' {
		out = append(out, '\n')
	}
	return out
}
