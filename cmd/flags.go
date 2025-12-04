package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterFlagsForPrefixWithOverrides registers Cobra flags for all configuration options
// whose keys start with the provided prefix. It binds each flag to Viper using the
// option's key. Flag names are derived from the key suffix by converting underscores
// to hyphens, unless an explicit override is provided in the overrides map.
// Returns an error if flag binding fails.
func RegisterFlagsForPrefixWithOverrides(cmd *cobra.Command, prefix string, overrides map[string]string) error {
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
			return fmt.Errorf("failed to bind flag %s to key %s: %w", flagName, opt.Key, err)
		}
	}

	return nil
}

func stringDefault(v interface{}) string {
	if v == nil {
		return ""
	}

	// Try direct string type
	if s, ok := v.(string); ok {
		return s
	}

	// Fallback to formatting (less efficient but handles edge cases)
	result := fmt.Sprintf("%v", v)
	log.Debug().
		Interface("value", v).
		Str("type", fmt.Sprintf("%T", v)).
		Str("result", result).
		Msg("Converting non-string value to string")

	return result
}

func boolDefault(v interface{}) bool {
	if v == nil {
		return false
	}

	// Try direct bool type
	if b, ok := v.(bool); ok {
		return b
	}

	// Try string conversion
	if s, ok := v.(string); ok {
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b
		}
		log.Warn().
			Str("value", s).
			Msg("Invalid string value for bool config, using false")
		return false
	}

	// Try numeric types (0 = false, non-zero = true)
	switch t := v.(type) {
	case int:
		if t != 0 {
			log.Debug().Interface("value", v).Msg("Converting non-zero numeric to true")
			return true
		}
		return false
	case int64:
		if t != 0 {
			log.Debug().Interface("value", v).Msg("Converting non-zero numeric to true")
			return true
		}
		return false
	case int32:
		if t != 0 {
			log.Debug().Interface("value", v).Msg("Converting non-zero numeric to true")
			return true
		}
		return false
	case int16:
		if t != 0 {
			log.Debug().Interface("value", v).Msg("Converting non-zero numeric to true")
			return true
		}
		return false
	case int8:
		if t != 0 {
			log.Debug().Interface("value", v).Msg("Converting non-zero numeric to true")
			return true
		}
		return false
	}

	log.Error().
		Interface("value", v).
		Str("type", fmt.Sprintf("%T", v)).
		Msg("Invalid type for bool config default, using false")

	return false
}

func intDefault(v interface{}) int {
	if v == nil {
		return 0
	}

	// Try int first
	if i, ok := v.(int); ok {
		return i
	}

	// Try other integer types
	switch t := v.(type) {
	case int64:
		// Check for overflow
		if t > int64(^uint(0)>>1) || t < -int64(^uint(0)>>1)-1 {
			log.Warn().
				Int64("value", t).
				Msg("Integer overflow in config default, clamping to max/min int")
			if t > 0 {
				return int(^uint(0) >> 1) // max int
			}
			return -int(^uint(0)>>1) - 1 // min int
		}
		return int(t)
	case int32:
		return int(t)
	case int16:
		return int(t)
	case int8:
		return int(t)
	case uint:
		// Check for overflow - uint can be larger than max int
		maxInt := ^uint(0) >> 1
		if t > maxInt {
			log.Warn().
				Uint("value", t).
				Msg("Unsigned integer overflow in config default, clamping to max int")
			return int(maxInt)
		}
		return int(t)
	case uint64:
		// Check for overflow - uint64 can be larger than max int
		maxInt := uint64(^uint(0) >> 1)
		if t > maxInt {
			log.Warn().
				Uint64("value", t).
				Msg("Unsigned integer overflow in config default, clamping to max int")
			return int(maxInt)
		}
		return int(t)
	case uint32:
		// Check for overflow - uint32 can be larger than max int on 32-bit systems
		// Convert to int64 first to check safely, then clamp if needed
		asInt64 := int64(t)
		maxInt := int64(^uint(0) >> 1)
		if asInt64 > maxInt {
			log.Warn().
				Uint32("value", t).
				Msg("Unsigned integer overflow in config default, clamping to max int")
			return int(maxInt)
		}
		return int(t)
	case uint16:
		return int(t)
	case uint8:
		return int(t)
	case float64:
		log.Warn().
			Float64("value", t).
			Msg("Float converted to int in config default, precision may be lost")
		return int(t)
	case float32:
		log.Warn().
			Float32("value", t).
			Msg("Float converted to int in config default, precision may be lost")
		return int(t)
	case string:
		// Try parsing string to int
		if i, err := strconv.Atoi(t); err == nil {
			return i
		}
		log.Error().
			Str("value", t).
			Msg("Invalid string value for int config, using 0")
		return 0
	default:
		log.Error().
			Interface("value", v).
			Str("type", fmt.Sprintf("%T", v)).
			Msg("Invalid type for int config default, using 0")
		return 0
	}
}

func floatDefault(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case int16:
		return float64(t)
	case int8:
		return float64(t)
	case uint:
		return float64(t)
	case uint64:
		return float64(t)
	case uint32:
		return float64(t)
	case uint16:
		return float64(t)
	case uint8:
		return float64(t)
	case string:
		// Try parsing string to float
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f
		}
		log.Error().
			Str("value", t).
			Msg("Invalid string value for float config, using 0.0")
		return 0
	default:
		log.Error().
			Interface("value", v).
			Str("type", fmt.Sprintf("%T", v)).
			Msg("Invalid type for float config default, using 0.0")
		return 0
	}
}

func stringSliceDefault(v interface{}) []string {
	if v == nil {
		return nil
	}

	// Try []string first
	if s, ok := v.([]string); ok {
		return s
	}

	// Try []interface{} (common from YAML/JSON parsing)
	if arr, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for i, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				log.Warn().
					Int("index", i).
					Interface("value", item).
					Str("type", fmt.Sprintf("%T", item)).
					Msg("Non-string item in array, converting to string")
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	}

	// Try single string (convert to array with one element)
	if s, ok := v.(string); ok {
		log.Debug().
			Str("value", s).
			Msg("Converting single string to string slice")
		return []string{s}
	}

	log.Error().
		Interface("value", v).
		Str("type", fmt.Sprintf("%T", v)).
		Msg("Invalid type for string slice config default, using empty slice")

	return nil
}
