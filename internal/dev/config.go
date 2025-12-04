// internal/dev/config.go
//
// Configuration inspector for development builds.
// Provides utilities to inspect, validate, and export configuration.

package dev

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/viper"
)

// ConfigInspector provides utilities for inspecting configuration
type ConfigInspector struct {
	registry []config.ConfigOption
}

// NewConfigInspector creates a new ConfigInspector
func NewConfigInspector() *ConfigInspector {
	return &ConfigInspector{
		registry: config.Registry(),
	}
}

// ListAllOptions returns all registered configuration options
func (ci *ConfigInspector) ListAllOptions() []config.ConfigOption {
	// Sort by key for consistent output
	sorted := make([]config.ConfigOption, len(ci.registry))
	copy(sorted, ci.registry)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})
	return sorted
}

// GetEffectiveConfig returns a map of all config keys with their effective values
func (ci *ConfigInspector) GetEffectiveConfig() map[string]interface{} {
	result := make(map[string]interface{})
	for _, opt := range ci.registry {
		result[opt.Key] = viper.Get(opt.Key)
	}
	return result
}

// GetConfigSources returns config values grouped by source
type ConfigSources struct {
	Defaults    map[string]interface{} `json:"defaults"`
	File        map[string]interface{} `json:"file,omitempty"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	Effective   map[string]interface{} `json:"effective"`
}

// GetConfigSourceInfo returns configuration values grouped by their source
func (ci *ConfigInspector) GetConfigSourceInfo(envPrefix string) *ConfigSources {
	sources := &ConfigSources{
		Defaults:    make(map[string]interface{}),
		File:        make(map[string]interface{}),
		Environment: make(map[string]interface{}),
		Effective:   make(map[string]interface{}),
	}

	for _, opt := range ci.registry {
		// Defaults
		sources.Defaults[opt.Key] = opt.DefaultValue

		// Effective (what Viper actually returns)
		sources.Effective[opt.Key] = viper.Get(opt.Key)

		// Check if value came from environment
		envName := opt.EnvVarName(envPrefix)
		if viper.IsSet(opt.Key) {
			// If value is set and differs from default, it's from file or env
			if fmt.Sprintf("%v", viper.Get(opt.Key)) != fmt.Sprintf("%v", opt.DefaultValue) {
				// Simple heuristic: if env var would bind this key, assume env source
				// (More sophisticated: check viper internals, but this is good enough for dev tool)
				if viper.GetString(envName) != "" {
					sources.Environment[opt.Key] = viper.Get(opt.Key)
				} else {
					sources.File[opt.Key] = viper.Get(opt.Key)
				}
			}
		}
	}

	return sources
}

// FormatAsTable returns a formatted table of configuration options
func (ci *ConfigInspector) FormatAsTable() string {
	var sb strings.Builder
	sb.WriteString("Configuration Registry:\n\n")
	sb.WriteString(fmt.Sprintf("%-30s %-15s %-20s %s\n", "KEY", "TYPE", "DEFAULT", "DESCRIPTION"))
	sb.WriteString(strings.Repeat("-", 100) + "\n")

	for _, opt := range ci.ListAllOptions() {
		defaultVal := opt.DefaultValueString()
		if len(defaultVal) > 18 {
			defaultVal = defaultVal[:15] + "..."
		}
		description := opt.Description
		if len(description) > 35 {
			description = description[:32] + "..."
		}

		sb.WriteString(fmt.Sprintf("%-30s %-15s %-20s %s\n",
			opt.Key,
			opt.Type,
			defaultVal,
			description,
		))
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d configuration options\n", len(ci.registry)))
	return sb.String()
}

// FormatEffectiveAsTable returns a formatted table of effective configuration
func (ci *ConfigInspector) FormatEffectiveAsTable() string {
	var sb strings.Builder
	sb.WriteString("Effective Configuration:\n\n")
	sb.WriteString(fmt.Sprintf("%-30s %-20s %s\n", "KEY", "VALUE", "SOURCE"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	effective := ci.GetEffectiveConfig()
	keys := make([]string, 0, len(effective))
	for k := range effective {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := fmt.Sprintf("%v", effective[key])
		if len(value) > 18 {
			value = value[:15] + "..."
		}

		// Determine source
		source := "default"
		for _, opt := range ci.registry {
			if opt.Key == key {
				if fmt.Sprintf("%v", effective[key]) != fmt.Sprintf("%v", opt.DefaultValue) {
					source = "file/env"
				}
				break
			}
		}

		sb.WriteString(fmt.Sprintf("%-30s %-20s %s\n", key, value, source))
	}

	return sb.String()
}

// ExportToJSON exports configuration to JSON format
func (ci *ConfigInspector) ExportToJSON(includeEffective bool) (string, error) {
	var data interface{}

	if includeEffective {
		data = struct {
			Registry  []config.ConfigOption  `json:"registry"`
			Effective map[string]interface{} `json:"effective"`
		}{
			Registry:  ci.ListAllOptions(),
			Effective: ci.GetEffectiveConfig(),
		}
	} else {
		data = ci.ListAllOptions()
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// ValidateConfig validates the current configuration
func (ci *ConfigInspector) ValidateConfig() []error {
	var errors []error

	for _, opt := range ci.registry {
		// Check required fields
		if opt.Required {
			value := viper.Get(opt.Key)
			if value == nil || fmt.Sprintf("%v", value) == "" {
				errors = append(errors, fmt.Errorf("required config key '%s' is not set", opt.Key))
			}
		}

		// TODO: Add type validation if needed
		// This would require type assertions based on opt.Type
	}

	return errors
}

// GetConfigByPrefix returns all config options matching a key prefix
func (ci *ConfigInspector) GetConfigByPrefix(prefix string) []config.ConfigOption {
	var matches []config.ConfigOption
	for _, opt := range ci.registry {
		if strings.HasPrefix(opt.Key, prefix) {
			matches = append(matches, opt)
		}
	}
	return matches
}
