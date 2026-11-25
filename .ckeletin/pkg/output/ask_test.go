package output

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsk_RefusesInJSONMode(t *testing.T) {
	SetOutputMode("json")
	t.Cleanup(func() { SetOutputMode("") })

	cmd := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("yes\n"))

	answer, err := Ask(cmd, "Delete it?", "--yes")
	require.Error(t, err)
	assert.Empty(t, answer)
	assert.Contains(t, err.Error(), "pass --yes")
	assert.Empty(t, errBuf.String())
}

// The streams report a terminal, so the only reason to refuse is JSON mode.
// Removing that check lets the prompt run and this test goes red.
func TestAsk_RefusesJSONModeEvenOnATerminal(t *testing.T) {
	SetOutputMode("json")
	t.Cleanup(func() { SetOutputMode("") })

	cmd := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("yes\n"))

	alwaysOut := func(io.Writer) bool { return true }
	alwaysIn := func(io.Reader) bool { return true }
	answer, err := ask(cmd, "Delete it?", "--yes", alwaysOut, alwaysIn)
	require.Error(t, err)
	assert.Empty(t, answer)
	assert.Contains(t, err.Error(), "pass --yes")
	assert.Empty(t, errBuf.String(), "JSON mode must not write the question")
}

func TestAsk_RefusesWhenTheStreamIsNotATerminal(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() { SetOutputMode("") })

	cmd := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("yes\n"))

	neverOut := func(io.Writer) bool { return false }
	neverIn := func(io.Reader) bool { return false }
	answer, err := ask(cmd, "Name?", "--name", neverOut, neverIn)
	require.Error(t, err)
	assert.Empty(t, answer)
	assert.Contains(t, err.Error(), "pass --name")
	assert.Empty(t, errBuf.String())
}

func TestAsk_ReadsOneLine(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() { SetOutputMode("") })

	cmd := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("alice\n"))

	alwaysOut := func(io.Writer) bool { return true }
	alwaysIn := func(io.Reader) bool { return true }
	answer, err := ask(cmd, "Name?", "--name", alwaysOut, alwaysIn)
	require.NoError(t, err)
	assert.Equal(t, "alice", answer)
	assert.Equal(t, "Name?\n", errBuf.String())
}

func TestAsk_NamesAnArgumentWhenNoneIsGiven(t *testing.T) {
	SetOutputMode("json")
	t.Cleanup(func() { SetOutputMode("") })

	answer, err := Ask(nil, "Name?\n", "  ")
	require.Error(t, err)
	assert.Empty(t, answer)
	assert.Contains(t, err.Error(), "pass the value as an argument")
	assert.Equal(t, os.Stdin, stdIn(nil))
	assert.False(t, readerIsTerminal(strings.NewReader("x")))
	_ = readerIsTerminal(os.Stdin)
}

func TestAsk_WriteErrorAndEmptyInput(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() { SetOutputMode("") })

	alwaysOut := func(io.Writer) bool { return true }
	alwaysIn := func(io.Reader) bool { return true }

	closed := &cobra.Command{Use: "app"}
	closed.SetErr(failWrite{})
	closed.SetIn(strings.NewReader("x\n"))
	_, err := ask(closed, "Name?\n", "--name", alwaysOut, alwaysIn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	empty := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	empty.SetErr(&errBuf)
	empty.SetIn(strings.NewReader(""))
	_, err = ask(empty, "Name?", "--name", alwaysOut, alwaysIn)
	require.Error(t, err)
}

type failWrite struct{}

func (failWrite) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestShadowLog_BoundsALongWrite(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() {
		SetOutputMode("")
		log.Logger = zerolog.New(io.Discard)
	})

	var audit bytes.Buffer
	log.Logger = zerolog.New(&audit)
	cmd := &cobra.Command{Use: "app"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	body := strings.Repeat("x", maxAuditOutputBytes+50)
	require.NoError(t, Resultf(cmd, "%s", body))
	assert.Equal(t, body, out.String())
	assert.Contains(t, audit.String(), `"truncated":true`)
	assert.NotContains(t, audit.String(), strings.Repeat("x", maxAuditOutputBytes+1))
}
