package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
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
		return nil, maskAnyf(invalidConfigError, "description commit must not be empty")
	}
	if config.GitCommit == "" {
		return nil, maskAnyf(invalidConfigError, "git commit must not be empty")
	}
	if config.Name == "" {
		return nil, maskAnyf(invalidConfigError, "name must not be empty")
	}
	if config.ProjectVersion == "" {
		return nil, maskAnyf(invalidConfigError, "project version must not be empty")
	}
	if config.Source == "" {
		return nil, maskAnyf(invalidConfigError, "name must not be empty")
	}

	newCommand := &command{
		Config: config,
	}

	return newCommand, nil
}

// command represents the version command.
type command struct {
	Config
}

// Execute represents the cobra run method.
func (c *command) Execute(cmd *cobra.Command, args []string) {
	fmt.Printf("Description:        %s\n", c.Description)
	fmt.Printf("Git Commit:         %s\n", c.GitCommit)
	fmt.Printf("Go Version:         %s\n", runtime.Version())
	fmt.Printf("Name:               %s\n", c.Name)
	fmt.Printf("OS / Arch:          %s / %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Project Version:    %s\n", c.ProjectVersion)
	fmt.Printf("Source:             %s\n", c.Source)
}

// New creates a new cobra command for the version command.
func (c *command) New() *cobra.Command {
	newCommand := &cobra.Command{
		Use:   "version",
		Short: "Show version information of the microservice.",
		Long:  "Show version information of the microservice.",
		Run:   c.Execute,
	}

	return newCommand
}
