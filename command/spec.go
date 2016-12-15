package command

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microkit/command/daemon"
	"github.com/giantswarm/microkit/command/version"
)

type Command interface {
	CobraCommand() *cobra.Command
	DaemonCommand() daemon.Command
	Execute(cmd *cobra.Command, args []string)
	VersionCommand() version.Command
}
