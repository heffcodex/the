package the

import (
	"github.com/spf13/cobra"
)

type CmdOption func(cmd *cobra.Command)

func SilenceErrors(v bool) CmdOption {
	return func(cmd *cobra.Command) {
		cmd.SilenceErrors = v
	}
}

func SilenceUsage(v bool) CmdOption {
	return func(cmd *cobra.Command) {
		cmd.SilenceUsage = v
	}
}
