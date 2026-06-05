package catalog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCatalogJSONShape pins the cross-implementation contract: the required core
// is always present (even a false bool), optional fields are omitted when empty,
// and empty slices render as [] not null — so a single agent parser works across
// ckeletin-go and ckeletin-rust.
func TestCatalogJSONShape(t *testing.T) {
	cat := Catalog{
		Name:        "demo",
		Description: "A demo CLI",
		GlobalFlags: []Flag{{
			Long:        "output",
			Required:    false,
			TakesValue:  true,
			Description: "Output format",
			Default:     "text",
		}},
		Commands: []Command{{
			Name:        "ping",
			Description: "Check connectivity",
			Flags:       []Flag{},
			Commands:    []Command{},
		}},
	}

	b, err := json.Marshal(cat)
	require.NoError(t, err)
	s := string(b)

	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))

	// Required core present.
	assert.Equal(t, "demo", m["name"])
	gf := m["global_flags"].([]any)[0].(map[string]any)
	assert.Equal(t, "output", gf["long"])
	assert.Equal(t, true, gf["takes_value"])
	// required:false MUST still serialize (no omitempty) — the cross-impl core.
	require.Contains(t, gf, "required")
	assert.Equal(t, false, gf["required"])

	// Optionals omitted when empty (matches rust serde skip_serializing_if).
	assert.NotContains(t, gf, "short")
	assert.NotContains(t, gf, "possible_values")

	// Empty slices serialize as [] not null (cross-impl shape parity).
	assert.Contains(t, s, `"flags":[]`)
	assert.Contains(t, s, `"commands":[]`)
}

func TestCatalogString(t *testing.T) {
	cat := Catalog{
		Name:        "demo",
		Description: "A demo CLI",
		GlobalFlags: []Flag{{Long: "output", TakesValue: true, Description: "Output format"}},
		Commands:    []Command{{Name: "ping", Description: "Check connectivity"}},
	}

	out := cat.String()
	assert.Contains(t, out, "Commands:")
	assert.Contains(t, out, "ping")
	assert.Contains(t, out, "--output <value>")
	assert.Contains(t, out, "Output format")
}
