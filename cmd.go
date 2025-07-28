package the

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/heffcodex/the/tcfg"
)

type NewAppFunc[A App[C], C tcfg.Config] func() (A, error)

type Cmd[A App[C], C tcfg.Config] struct {
	newApp NewAppFunc[A, C]
	opts   []CmdOption
}

func NewCmd[A App[C], C tcfg.Config](newApp NewAppFunc[A, C], opts ...CmdOption) *Cmd[A, C] {
	return &Cmd[A, C]{
		newApp: newApp,
		opts:   opts,
	}
}

func (c *Cmd[A, C]) Execute() error {
	const recoverStackSkip = 2

	defer func() {
		if e := recover(); e != nil {
			err := fmt.Errorf("%+v", e) //nolint:err113 // ok to construct dynamic err from panic value
			zap.L().Fatal("panic", zap.Error(err), zap.StackSkip("stack", recoverStackSkip))
		}
	}()

	shut := newShutter([]os.Signal{syscall.SIGINT, syscall.SIGTERM})

	root, err := c.makeRoot(shut)
	if err != nil {
		return err
	}

	if err = root.Execute(); err != nil {
		shut.down()
		return err
	}

	return nil
}

func (c *Cmd[A, C]) makeRoot(shut *shutter) (*cobra.Command, error) {
	app, err := c.newApp()
	if err != nil {
		return nil, fmt.Errorf("new app: %w", err)
	}

	root := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cancelFn := cmdInject[A, C](cmd, app, shut)
			timeout := app.C().ShutdownTimeout()

			shut.setup(app.L().Named("cmd"), cancelFn, app.Close, timeout)
			go func() {
				shut.rootWaitInterrupt()
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

	return root, nil
}
