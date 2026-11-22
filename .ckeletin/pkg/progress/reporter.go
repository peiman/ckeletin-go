// Package progress is the port business logic uses to report work.
// It has no stream, format, or terminal knowledge. The command layer
// supplies an implementation. Principle 8: the framework defines the
// edge, and each application does not invent its own.
package progress

// Reporter is injected into business functions. Methods do not write
// stdout or stderr. The command layer decides where the words go.
type Reporter interface {
	Step(message string)
	Status(message string)
	Warn(message string)
	Done(message string)
}

// Nop discards every report. A nil reporter is treated as Nop.
type Nop struct{}

func (Nop) Step(string)   {}
func (Nop) Status(string) {}
func (Nop) Warn(string)   {}
func (Nop) Done(string)   {}

// Event is one call recorded by Recorder.
type Event struct {
	Kind    string
	Message string
}

// Recorder is the test double. It stores calls and writes nothing.
type Recorder struct {
	Events []Event
}

func (r *Recorder) Step(message string)   { r.add("step", message) }
func (r *Recorder) Status(message string) { r.add("status", message) }
func (r *Recorder) Warn(message string)   { r.add("warn", message) }
func (r *Recorder) Done(message string)   { r.add("done", message) }

func (r *Recorder) add(kind, message string) {
	r.Events = append(r.Events, Event{Kind: kind, Message: message})
}
