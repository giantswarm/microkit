package command

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microkit/command/daemon"
	"github.com/giantswarm/microkit/command/version"
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
			return nil, maskAny(err)
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
			return nil, maskAny(err)
		}
	}

	newCommand := &command{
		// Dependencies.
		DaemonCommand:  daemonCommand,
		VersionCommand: versionCommand,

		// Settings.
		Name:        config.Name,
		Description: config.Description,
	}

	return newCommand, nil
}

// command represents the root command.
type command struct {
	// Dependencies.
	DaemonCommand  daemon.Command
	VersionCommand version.Command

	// Settings.
	Name        string
	Description string
}

// Execute represents the cobra run method.
func (c *command) Execute(cmd *cobra.Command, args []string) {
	cmd.HelpFunc()(cmd, nil)
}

// New creates a new cobra command for the root command.
func (c *command) New() *cobra.Command {
	newCommand := &cobra.Command{
		Use:   c.Name,
		Short: c.Description,
		Long:  c.Description,
		Run:   c.Execute,
	}

	newCommand.AddCommand(c.DaemonCommand.New())
	newCommand.AddCommand(c.VersionCommand.New())

	return newCommand
}
