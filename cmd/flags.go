package cmd

import (
	"fmt"
	"strings"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterFlagsForPrefixWithOverrides registers Cobra flags for all configuration options
// whose keys start with the provided prefix. It binds each flag to Viper using the
// option's key. Flag names are derived from the key suffix by converting underscores
// to hyphens, unless an explicit override is provided in the overrides map.
func RegisterFlagsForPrefixWithOverrides(cmd *cobra.Command, prefix string, overrides map[string]string) {
	options := config.Registry()

	for _, opt := range options {
		if !strings.HasPrefix(opt.Key, prefix) {
			continue
		}

		// Derive default flag name from key suffix
		suffix := strings.TrimPrefix(opt.Key, prefix)
		defaultFlag := strings.ReplaceAll(suffix, "_", "-")

		flagName := defaultFlag
		if overrides != nil {
			if custom, ok := overrides[opt.Key]; ok {
				flagName = custom
			}
		}

		// Create flag based on option type and bind to Viper
		switch strings.ToLower(opt.Type) {
		case "string":
			cmd.Flags().String(flagName, stringDefault(opt.DefaultValue), opt.Description)
		case "bool":
			cmd.Flags().Bool(flagName, boolDefault(opt.DefaultValue), opt.Description)
		case "int":
			cmd.Flags().Int(flagName, intDefault(opt.DefaultValue), opt.Description)
		case "float", "float64":
			cmd.Flags().Float64(flagName, floatDefault(opt.DefaultValue), opt.Description)
		case "[]string", "stringslice":
			cmd.Flags().StringSlice(flagName, stringSliceDefault(opt.DefaultValue), opt.Description)
		default:
			// Fallback to string flag if type is unknown, but log a warning
			log.Warn().Str("key", opt.Key).Str("type", opt.Type).Msg("Unknown option type, defaulting to string flag")
			cmd.Flags().String(flagName, stringDefault(opt.DefaultValue), opt.Description)
		}

		if err := viper.BindPFlag(opt.Key, cmd.Flags().Lookup(flagName)); err != nil {
			log.Fatal().Err(err).Str("key", opt.Key).Str("flag", flagName).Msg("Failed to bind flag")
		}
	}
}

func stringDefault(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func boolDefault(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func intDefault(v interface{}) int {
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}

func floatDefault(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

func stringSliceDefault(v interface{}) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}
