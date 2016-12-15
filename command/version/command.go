package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	microerror "github.com/giantswarm/microkit/error"
)

// Config represents the configuration used to create a new version command.
type Config struct {
	// Settings.
	Description    string
	GitCommit      string
	Name           string
	ProjectVersion string
	Source         string
}

// DefaultConfig provides a default configuration to create a new version
// command by best effort.
func DefaultConfig() Config {
	return Config{
		// Settings.
		Description:    "",
		GitCommit:      "",
		Name:           "",
		ProjectVersion: "",
		Source:         "",
	}
}

// New creates a new configured version command.
func New(config Config) (Command, error) {
	// Settings.
	if config.Description == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "description commit must not be empty")
	}
	if config.GitCommit == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "git commit must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "name must not be empty")
	}
	if config.ProjectVersion == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "project version must not be empty")
	}
	if config.Source == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "name must not be empty")
	}

	newCommand := &command{
		// Internals.
		cobraCommand: nil,

		// Settings.
		description:    config.Description,
		gitCommit:      config.GitCommit,
		name:           config.Name,
		projectVersion: config.ProjectVersion,
		source:         config.Source,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   "version",
		Short: "Show version information of the microservice.",
		Long:  "Show version information of the microservice.",
		Run:   newCommand.Execute,
	}

	return newCommand, nil
}

// command represents the version command.
type command struct {
	// Internals.
	cobraCommand *cobra.Command

	// Settings.
	description    string
	gitCommit      string
	name           string
	projectVersion string
	source         string
}

// CobraCommand returns the actual cobra command for the version command.
func (c *command) CobraCommand() *cobra.Command {
	return c.cobraCommand
}

// Execute represents the cobra run method.
func (c *command) Execute(cmd *cobra.Command, args []string) {
	fmt.Printf("Description:        %s\n", c.description)
	fmt.Printf("Git Commit:         %s\n", c.gitCommit)
	fmt.Printf("Go Version:         %s\n", runtime.Version())
	fmt.Printf("Name:               %s\n", c.name)
	fmt.Printf("OS / Arch:          %s / %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Project Version:    %s\n", c.projectVersion)
	fmt.Printf("Source:             %s\n", c.source)
}
