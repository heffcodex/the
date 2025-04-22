package the

import (
	"github.com/spf13/cobra"

	"github.com/heffcodex/the/tcfg"
)

type CmdOption func(cmd *cobra.Command)

func CmdOptionFunc(fn func(cmd *cobra.Command)) CmdOption {
	return fn
}

func SilenceAll() CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		SilenceErrors(true)(cmd)
		SilenceUsage(true)(cmd)
	})
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

func OnAppReady[A App[C], C tcfg.Config](fns ...func(app A) error) CmdOption {
	return CmdOptionFunc(func(cmd *cobra.Command) {
		preRunE := cmd.PersistentPreRunE

		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := preRunE(cmd, args); err != nil {
				return err
			}

			app := CmdApp[A, C](cmd)

			for _, fn := range fns {
				if err := fn(app); err != nil {
					return err
				}
			}

			return nil
		}
	})
}
