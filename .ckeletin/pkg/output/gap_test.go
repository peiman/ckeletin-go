package output

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONMode_RecordsStatusAndResult(t *testing.T) {
	SetOutputMode("json")
	t.Cleanup(func() {
		SetOutputMode("")
		log.Logger = zerolog.New(ioDiscard)
	})

	var audit bytes.Buffer
	log.Logger = zerolog.New(&audit)
	cmd := &cobra.Command{Use: "app"}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	require.NoError(t, Statusf(cmd, "quality warning"))
	require.NoError(t, Resultf(cmd, "plan text"))
	assert.Empty(t, out.String())
	assert.Empty(t, errBuf.String())
	assert.Contains(t, audit.String(), "quality warning")
	assert.Contains(t, audit.String(), "plan text")
	assert.Contains(t, audit.String(), `"stream":"stderr"`)
	assert.Contains(t, audit.String(), `"stream":"stdout"`)
}

var ioDiscard = discardWriter{}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestPanic_RecordsTheStack(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() {
		SetOutputMode("")
		log.Logger = zerolog.New(ioDiscard)
		SetPrepare(nil)
	})

	var audit bytes.Buffer
	log.Logger = zerolog.New(&audit)
	root := &cobra.Command{Use: "app"}
	root.RunE = func(*cobra.Command, []string) error {
		panic("boom")
	}
	var errBuf bytes.Buffer
	root.SetErr(&errBuf)
	root.SetOut(&bytes.Buffer{})

	code := Execute(root)
	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "panic: boom")
	assert.Contains(t, audit.String(), "recovered panic")
	assert.Contains(t, audit.String(), "goroutine")
	assert.NotContains(t, errBuf.String(), "goroutine")
}
