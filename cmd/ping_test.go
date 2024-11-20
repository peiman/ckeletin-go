// cmd/ping_test.go - Tests for the ping command
package cmd

import (
	"bytes"
	"testing"

	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

// PingCommandSuite defines a test suite for the ping command
type PingCommandSuite struct {
	suite.Suite
	output *bytes.Buffer
}

func (s *PingCommandSuite) SetupTest() {
	// Reset viper config for each test
	viper.Reset()
	viper.SetDefault("ping.defaultCount", infrastructure.DefaultPingCount)
	viper.SetDefault("ping.outputMessage", infrastructure.DefaultPingMessage)
	viper.SetDefault("ping.coloredOutput", false)

	// Setup logger
	err := infrastructure.InitLogger("debug")
	s.Require().NoError(err, "Failed to initialize logger")

	s.output = &bytes.Buffer{}
}

func TestPingCommandSuite(t *testing.T) {
	suite.Run(t, new(PingCommandSuite))
}

func (s *PingCommandSuite) TestBasicPing() {
	// Given
	cmd := newPingCommand()
	cmd.SetOut(s.output)

	// When
	err := cmd.Execute()

	// Then
	s.Require().NoError(err)
	s.Equal("pong\n", s.output.String())
}

func (s *PingCommandSuite) TestPingWithCustomConfig() {
	// Given
	viper.Set("ping.defaultCount", 2)
	viper.Set("ping.outputMessage", "hello")
	cmd := newPingCommand()
	cmd.SetOut(s.output)

	// When
	err := cmd.Execute()

	// Then
	s.Require().NoError(err)
	s.Equal("hello\nhello\n", s.output.String())
}

func (s *PingCommandSuite) TestPingFlagOverridesConfig() {
	// Given
	viper.Set("ping.defaultCount", 5)
	cmd := newPingCommand()
	cmd.SetOut(s.output)
	cmd.SetArgs([]string{"--count", "2"})

	// When
	err := cmd.Execute()

	// Then
	s.Require().NoError(err)
	s.Equal("pong\npong\n", s.output.String())
}

func (s *PingCommandSuite) TestPingWithInvalidCount() {
	// Given
	cmd := newPingCommand()
	cmd.SetOut(s.output)
	cmd.SetArgs([]string{"--count", "-1"})

	// When
	err := cmd.Execute()

	// Then
	s.Require().Error(err)
	s.Contains(err.Error(), "count flag must be greater than 0")
}

func (s *PingCommandSuite) TestPingCommandFlags() {
	// Given
	cmd := newPingCommand()

	// When
	countFlag := cmd.Flags().Lookup("count")

	// Then
	s.Require().NotNil(countFlag, "count flag should exist")
	s.Equal("1", countFlag.DefValue)
	s.Equal("number of times to ping", countFlag.Usage)
}
