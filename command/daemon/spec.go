package daemon

import (
	"github.com/spf13/cobra"
)

type Command interface {
	CobraCommand() *cobra.Command
	Execute(cmd *cobra.Command, args []string)
}
