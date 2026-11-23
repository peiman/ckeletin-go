package output

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Log level flags. A changed flag wins over the environment variable, which
// wins over the config file, which wins over the default. Viper applies that
// order once the flag is bound. The environment names are
// <PREFIX>_APP_LOG_LEVEL, <PREFIX>_APP_LOG_CONSOLE_LEVEL, and
// <PREFIX>_APP_LOG_FILE_LEVEL. The reference prefix is CKELETIN_GO.
const (
	logLevelFlag        = "log-level"
	logLevelConsoleFlag = "log-level-console"
	logLevelFileFlag    = "log-level-file"
)

// RegisterLogLevelFlags registers persistent --log-level, --log-level-console,
// and --log-level-file, bound to app.log_level, app.log.console_level, and
// app.log.file_level. The bind runs at the start of PersistentPreRunE, before
// the application's own pre-run, so logger.Init sees the flag.
func RegisterLogLevelFlags(root *cobra.Command) {
	if root == nil {
		return
	}
	flags := root.PersistentFlags()
	if flags.Lookup(logLevelFlag) == nil {
		flags.String(logLevelFlag, "warn",
			"Console log level when --log-level-console is empty (trace, debug, info, warn, error, fatal, panic). Does not change the file level.")
	}
	if flags.Lookup(logLevelConsoleFlag) == nil {
		flags.String(logLevelConsoleFlag, "",
			"Console log level (trace, debug, info, warn, error, fatal, panic). Empty uses --log-level.")
	}
	if flags.Lookup(logLevelFileFlag) == nil {
		flags.String(logLevelFileFlag, "debug",
			"Audit file log level (trace, debug, info, warn, error, fatal, panic). Empty uses debug. Does not follow --log-level.")
	}
	if root.Annotations == nil {
		root.Annotations = map[string]string{}
	}
	if root.Annotations["ckeletin.loglevel.wrapped"] == "1" {
		return
	}
	root.Annotations["ckeletin.loglevel.wrapped"] = "1"

	prev := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := bindLogLevelFlags(cmd); err != nil {
			return err
		}
		if prev != nil {
			return prev(cmd, args)
		}
		return nil
	}
}

func bindLogLevelFlags(cmd *cobra.Command) error {
	if cmd == nil || cmd.Root() == nil {
		return fmt.Errorf("log level flags: no command")
	}
	flags := cmd.Root().PersistentFlags()
	pairs := []struct {
		key  string
		name string
	}{
		{config.KeyAppLogLevel, logLevelFlag},
		{config.KeyAppLogConsoleLevel, logLevelConsoleFlag},
		{config.KeyAppLogFileLevel, logLevelFileFlag},
	}
	for _, pair := range pairs {
		flag := flags.Lookup(pair.name)
		if flag == nil {
			return fmt.Errorf("log level flag %q is not registered", pair.name)
		}
		if err := viper.BindPFlag(pair.key, flag); err != nil {
			return fmt.Errorf("bind flag %q to %s: %w", pair.name, pair.key, err)
		}
	}
	return nil
}
