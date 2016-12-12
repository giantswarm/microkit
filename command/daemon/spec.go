package daemon

import (
	"github.com/spf13/cobra"
)

type Command interface {
	Execute(cmd *cobra.Command, args []string)
	New() *cobra.Command
}
