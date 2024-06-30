package the

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/heffcodex/the/tcfg"
)

type (
	appKey     struct{}
	shutterKey struct{}
)

func CApp[C tcfg.Config, A App[C]](cmd *cobra.Command) A {
	return getApp[C, A](cmd)
}

func getApp[C tcfg.Config, A App[C]](cmd *cobra.Command) A {
	return cmd.Context().Value(appKey{}).(A) //nolint: revive // unchecked-type-assertion: it's ok to panic here
}

func CWaitInterrupt(cmd *cobra.Command) {
	getShutter(cmd).waitInterrupt()
}

func getShutter(cmd *cobra.Command) *shutter {
	return cmd.Context().Value(shutterKey{}).(*shutter) //nolint: revive // unchecked-type-assertion: it's ok to panic here
}

func cmdInject[C tcfg.Config, A App[C]](cmd *cobra.Command, app A, shut *shutter) (cancel context.CancelFunc) {
	ctx := cmd.Context()

	ctx = context.WithValue(ctx, appKey{}, app)
	ctx = context.WithValue(ctx, shutterKey{}, shut)

	ctx, cancel = context.WithCancel(ctx)
	cmd.SetContext(ctx)

	return cancel
}
