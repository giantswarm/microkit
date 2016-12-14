package command

import (
	"github.com/spf13/cobra"
)

type Command interface {
	CobraCommand() *cobra.Command
	DaemonCommand() *cobra.Command
	Execute(cmd *cobra.Command, args []string)
	VersionCommand() *cobra.Command
}
