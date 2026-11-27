package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEmptyFileLevelIsValid(t *testing.T) {
	SetDefaults()
	t.Cleanup(func() {
		viper.Set(KeyAppLogFileLevel, "debug")
	})
	viper.Set(KeyAppLogFileLevel, "")
	for _, err := range ValidateRegisteredOptions() {
		assert.NotContains(t, err.Error(), KeyAppLogFileLevel)
	}
}
