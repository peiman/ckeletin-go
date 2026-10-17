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
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	parsed := viper.New()
	parsed.SetConfigType("toml")
	if err := parsed.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("parse project defaults: %w", err)
	}

	for _, key := range parsed.AllKeys() {
		if key != "defaults" && !strings.HasPrefix(key, "defaults.") {
			return fmt.Errorf("project defaults: unknown table %q", key)
		}
	}

	sub := parsed.Sub("defaults")
	if sub == nil {
		return nil
	}

	known := make(map[string]struct{}, len(Registry()))
	for _, opt := range Registry() {
		known[opt.Key] = struct{}{}
	}

	keys := sub.AllKeys()
	for _, key := range keys {
		if _, ok := known[key]; !ok {
			return fmt.Errorf("project defaults: unknown option %q", key)
		}
	}
	for _, key := range keys {
		viper.SetDefault(key, sub.Get(key))
	}
	return nil
}
