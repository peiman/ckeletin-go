package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecorderStoresEachKind(t *testing.T) {
	var rec Recorder
	var rep Reporter = &rec
	rep.Step("one")
	rep.Status("two")
	rep.Warn("three")
	rep.Done("four")
	assert.Equal(t, []Event{
		{Kind: "step", Message: "one"},
		{Kind: "status", Message: "two"},
		{Kind: "warn", Message: "three"},
		{Kind: "done", Message: "four"},
	}, rec.Events)
}

func TestNopDiscards(t *testing.T) {
	var rep Reporter = Nop{}
	rep.Step("silent")
	rep.Done("silent")
}
