package the

import (
	"github.com/spf13/cobra"
)

type CmdOption func(cmd *cobra.Command)

func CmdOptionFunc(fn func(cmd *cobra.Command)) CmdOption {
	return fn
}

func SilenceErrors(v bool) CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		cmd.SilenceErrors = v
	})
}

func SilenceUsage(v bool) CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		cmd.SilenceUsage = v
	})
}

func Commands(commands ...*cobra.Command) CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		for _, add := range commands {
			cmd.AddCommand(add)
		}
	})
}

func Args(args ...string) CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		cmd.SetArgs(args)
	})
}
