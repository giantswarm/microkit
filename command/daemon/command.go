package daemon

import (
	"os"
	"os/signal"
	"sync"

	kitlog "github.com/go-kit/kit/log"
	"github.com/spf13/cobra"

	"github.com/giantswarm/microkit/logger"
	"github.com/giantswarm/microkit/server"
)

// Config represents the configuration used to create a new daemon command.
type Config struct {
	// Dependencies.
	Logger        logger.Logger
	ServerFactory func() server.Server
}

// DefaultConfig provides a default configuration to create a new daemon command
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:        nil,
		ServerFactory: nil,
	}
}

// New creates a new daemon command.
func New(config Config) (Command, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, maskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.ServerFactory == nil {
		return nil, maskAnyf(invalidConfigError, "server factory must not be empty")
	}

	config.Logger = kitlog.NewContext(config.Logger).With("package", "command/daemon")

	newCommand := &command{
		Config: config,
	}

	return newCommand, nil
}

// command represents the daemon command.
type command struct {
	Config
}

// Execute represents the cobra run method.
func (c *command) Execute(cmd *cobra.Command, args []string) {
	// Merge the given command line flags with the given environment variables and
	// the given config file, if any. The merged flags will be applied to the
	// global Flags struct.
	err := MergeFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}

	var newServer server.Server
	{
		customServer := c.ServerFactory()

		serverConfig := server.DefaultConfig()

		serverConfig.Endpoints = customServer.Endpoints()
		serverConfig.ErrorEncoder = customServer.ErrorEncoder()
		serverConfig.ListenAddress = Flags.Server.Listen.Address
		serverConfig.Logger = c.Logger
		serverConfig.RequestFuncs = customServer.RequestFuncs()

		newServer, err = server.New(serverConfig)
		if err != nil {
			panic(err)
		}
		go newServer.Boot()
	}

	// Listen to OS signals.
	listener := make(chan os.Signal, 2)
	signal.Notify(listener, os.Interrupt, os.Kill)

	<-listener

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			newServer.Shutdown()
		}()

		os.Exit(0)
	}()

	<-listener

	os.Exit(0)
}

// New creates a new cobra command for the daemon command.
func (c *command) New() *cobra.Command {
	newCommand := &cobra.Command{
		Use:   "daemon",
		Short: "Execute the daemon of the microservice.",
		Long:  "Execute the daemon of the microservice.",
		Run:   c.Execute,
	}

	newCommand.PersistentFlags().StringVar(&Flags.Config.Dir, "config.dir", ".", "Directory of the config file.")
	newCommand.PersistentFlags().StringVar(&Flags.Config.File, "config.file", "config", "Name of the config file. All viper supported extensions can be used.")

	newCommand.PersistentFlags().StringVar(&Flags.Server.Listen.Address, "server.listen.address", "http://127.0.0.1:8080", "Address used to make the server listen to.")

	return newCommand
}
