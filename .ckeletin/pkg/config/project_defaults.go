// .ckeletin/pkg/config/project_defaults.go
//
// Project overlay for framework defaults.
//
// The registry is the source of which options exist and of the framework
// default. A derived project may replace a default from its own TOML file.
// That file lives outside .ckeletin/. A config file, environment variable,
// or flag still wins, because this only calls viper.SetDefault.

package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// ApplyProjectDefaultsTOML replaces framework defaults with the values in a
// project TOML document. data is the file contents, not a path. An empty
// document or an empty [defaults] table changes nothing. A key the registry
// does not define returns an error and does not change any default.
func ApplyProjectDefaultsTOML(data []byte) error {
	overlay, err := parseProjectDefaults(data)
	if err != nil {
		return err
	}
	if err := validateProjectDefaults(overlay); err != nil {
		return err
	}
	for key, value := range overlay {
		viper.SetDefault(key, value)
	}
	return nil
}

// EffectiveRegistry returns the registry with project defaults applied to
// DefaultValue. The process defaults and the slices Registry returns are
// left unchanged. Generators use this so published defaults match the
// process. An empty document returns the framework defaults.
func EffectiveRegistry(data []byte) ([]ConfigOption, error) {
	overlay, err := parseProjectDefaults(data)
	if err != nil {
		return nil, err
	}
	if err := validateProjectDefaults(overlay); err != nil {
		return nil, err
	}

	opts := append([]ConfigOption(nil), Registry()...)
	if len(overlay) == 0 {
		return opts, nil
	}
	index := make(map[string]int, len(opts))
	for i, opt := range opts {
		index[opt.Key] = i
	}
	for key, value := range overlay {
		opts[index[key]].DefaultValue = value
	}
	return opts, nil
}

func parseProjectDefaults(data []byte) (map[string]interface{}, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	parsed := viper.New()
	parsed.SetConfigType("toml")
	if err := parsed.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parse project defaults: %w", err)
	}

	for _, key := range parsed.AllKeys() {
		if key != "defaults" && !strings.HasPrefix(key, "defaults.") {
			return nil, fmt.Errorf("project defaults: unknown table %q", key)
		}
	}

	sub := parsed.Sub("defaults")
	if sub == nil {
		return nil, nil
	}

	overlay := make(map[string]interface{}, len(sub.AllKeys()))
	for _, key := range sub.AllKeys() {
		overlay[key] = sub.Get(key)
	}
	return overlay, nil
}

func validateProjectDefaults(overlay map[string]interface{}) error {
	if len(overlay) == 0 {
		return nil
	}
	known := make(map[string]struct{}, len(Registry()))
	for _, opt := range Registry() {
		known[opt.Key] = struct{}{}
	}
	for key := range overlay {
		if _, ok := known[key]; !ok {
			return fmt.Errorf("project defaults: unknown option %q", key)
		}
	}
	return nil
}
