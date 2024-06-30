package the

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/heffcodex/the/tcfg"
)

type NewAppFunc[C tcfg.Config, A App[C]] func() (A, error)

type Cmd[C tcfg.Config, A App[C]] struct {
	newApp   NewAppFunc[C, A]
	opts     []CmdOption
	commands []*cobra.Command
}

func NewCmd[C tcfg.Config, A App[C]](newApp NewAppFunc[C, A], opts ...CmdOption) *Cmd[C, A] {
	return &Cmd[C, A]{
		newApp: newApp,
		opts:   opts,
	}
}

func (c *Cmd[C, A]) Add(commands ...*cobra.Command) {
	c.commands = append(c.commands, commands...)
}

func (c *Cmd[C, A]) Execute() error {
	const recoverStackSkip = 1

	defer func() {
		if e := recover(); e != nil {
			err := fmt.Errorf("%+v", e) //nolint: err113 // ok to construct dynamic err from panic value
			zap.L().Fatal("panic", zap.Error(err), zap.StackSkip("stack", recoverStackSkip))
		}
	}()

	shut := newShutter()
	root := c.makeRoot(shut)

	if err := root.Execute(); err != nil {
		shut.down()
		return err
	}

	return nil
}

func (c *Cmd[C, A]) makeRoot(shut *shutter) *cobra.Command {
	root := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			app, err := c.newApp()
			if err != nil {
				return fmt.Errorf("new app: %w", err)
			}

			cancelFn := cmdInject[C, A](cmd, app, shut)
			timeout := app.C().ShutdownTimeout()

			shut.setup(app.L(), cancelFn, app.Close, timeout)
			go func() {
				shut.waitInterrupt()
				shut.cancel()
			}()

			return nil
		},
		PersistentPostRun: func(*cobra.Command, []string) {
			shut.down()
		},
	}

	for _, opt := range c.opts {
		opt(root)
	}

	for _, cmd := range c.commands {
		root.AddCommand(cmd)
	}

	return root
}
