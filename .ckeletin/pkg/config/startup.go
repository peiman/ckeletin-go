package config

// StartupCheck is how the reference root treats a value or a config file
// that fails validation. Strict fails the command. Warn records the failure
// and continues, so a product can keep files its users already have.
type StartupCheck int

const (
	StartupStrict StartupCheck = iota
	StartupWarn
)

var startupCheck = StartupStrict

// SetStartupCheck selects strict or warn for the reference root. The default
// is strict. Call it before Execute.
func SetStartupCheck(check StartupCheck) {
	startupCheck = check
}

// StartupWarns reports whether a validation failure should be recorded and
// not fail the command.
func StartupWarns() bool {
	return startupCheck == StartupWarn
}
