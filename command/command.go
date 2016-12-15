package command

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microkit/command/daemon"
	"github.com/giantswarm/microkit/command/version"
	microerror "github.com/giantswarm/microkit/error"
	"github.com/giantswarm/microkit/logger"
	"github.com/giantswarm/microkit/server"
)

// Config represents the configuration used to create a new root command.
type Config struct {
	// Dependencies.
	Logger        logger.Logger
	ServerFactory func() server.Server

	// Settings.
	Description    string
	GitCommit      string
	Name           string
	ProjectVersion string
	Source         string
}

// DefaultConfig provides a default configuration to create a new root command
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:        nil,
		ServerFactory: nil,

		// Settings.
		Description:    "",
		GitCommit:      "",
		Name:           "",
		ProjectVersion: "",
		Source:         "",
	}
}

// New creates a new root command.
func New(config Config) (Command, error) {
	var err error

	var daemonCommand daemon.Command
	{
		daemonConfig := daemon.DefaultConfig()

		daemonConfig.Logger = config.Logger
		daemonConfig.ServerFactory = config.ServerFactory

		daemonCommand, err = daemon.New(daemonConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var versionCommand version.Command
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.Name
		versionConfig.ProjectVersion = config.ProjectVersion
		versionConfig.Source = config.Source

		versionCommand, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	newCommand := &command{
		// Internals.
		cobraCommand:   nil,
		daemonCommand:  daemonCommand,
		versionCommand: versionCommand,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   config.Name,
		Short: config.Description,
		Long:  config.Description,
		Run:   newCommand.Execute,
	}
	newCommand.cobraCommand.AddCommand(newCommand.daemonCommand.CobraCommand())
	newCommand.cobraCommand.AddCommand(newCommand.versionCommand.CobraCommand())

	return newCommand, nil
}

// command represents the root command.
type command struct {
	// Internals.
	cobraCommand   *cobra.Command
	daemonCommand  daemon.Command
	versionCommand version.Command
}

// CobraCommand returns the actual cobra command for the root command.
func (c *command) CobraCommand() *cobra.Command {
	return c.cobraCommand
}

// DaemonCommand returns the daemon sub command.
func (c *command) DaemonCommand() daemon.Command {
	return c.daemonCommand
}

// Execute represents the cobra run method.
func (c *command) Execute(cmd *cobra.Command, args []string) {
	cmd.HelpFunc()(cmd, nil)
}

// VersionCommand returns the version sub command.
func (c *command) VersionCommand() version.Command {
	return c.versionCommand
}
