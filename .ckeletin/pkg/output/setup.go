package output

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// WithAfterLogger runs after Setup has resolved the output mode.
// A nil error means the logger is installed and the failure audit is armed.
// A non-nil error means startup stopped first. The reference root replays
// or drops its pre-init buffer here, once the mode is known.
func WithAfterLogger(fn func(error)) Option {
	return func(s *flagSettings) {
		s.afterLogger = fn
	}
}

// Setup is the one call a derived root makes for machine-readable output.
//
// It registers --output, binds that flag to app.output_format, recognises an
// alias flag, and wraps PersistentPreRunE. The wrapper resolves the mode from
// the flag and the environment, runs the command's existing pre-run (that is
// where the config file is read), resolves the mode again, starts the logger,
// and records later failures on the audit log. Execute closes the log.
//
// The existing pre-run must load configuration and must not call logger.Init,
// UseOutputMode, SetFailureLog, or SetAfterRun. Setup does those, in that order.
func Setup(root *cobra.Command, opts ...Option) {
	if root == nil {
		return
	}
	RegisterOutputFlag(root, opts...)
	RegisterLogLevelFlags(root)
	logger.SetEnvPrefix(settings.envPrefix)
	hook := settings.afterLogger
	// The hook belongs to this root. A later Setup or RegisterOutputFlag
	// must not inherit it, and must not replace the closure captured below.
	settings.afterLogger = nil
	SetAfterRun(func() {
		logger.Cleanup()
		SetFailureLog(nil)
	})
	if root.Annotations != nil && root.Annotations["ckeletin.output.setup"] == "1" {
		return
	}
	if root.Annotations == nil {
		root.Annotations = map[string]string{}
	}
	root.Annotations["ckeletin.output.setup"] = "1"

	prev := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := bindLogLevelFlags(cmd); err != nil {
			return err
		}
		var runErr error
		if prev != nil {
			runErr = prev(cmd, args)
		}
		if runErr != nil {
			if modeErr := UseOutputMode(PhaseResolved); modeErr != nil {
				callAfterLogger(hook, modeErr)
				return modeErr
			}
			callAfterLogger(hook, runErr)
			return runErr
		}
		if err := UseOutputMode(PhaseResolved); err != nil {
			callAfterLogger(hook, err)
			return err
		}
		if err := logger.Init(nil); err != nil {
			wrapped := fmt.Errorf("failed to initialize logger: %w", err)
			callAfterLogger(hook, wrapped)
			return wrapped
		}
		installFailureLog()
		callAfterLogger(hook, nil)
		return nil
	}
}

func installFailureLog() {
	SetFailureLog(func(err error, command string, code int) {
		log.Debug().
			Err(err).
			Str("command", command).
			Int("exit_code", code).
			Msg("command failed")
	})
}

func callAfterLogger(hook func(error), err error) {
	if hook == nil {
		return
	}
	hook(err)
}
