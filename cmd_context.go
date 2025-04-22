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

func CmdApp[A App[C], C tcfg.Config](cmd *cobra.Command) A {
	return ContextApp[A, C](cmd.Context())
}

func ContextApp[A App[C], C tcfg.Config](ctx context.Context) A {
	return ctx.Value(appKey{}).(A) //nolint:errcheck,revive // it's ok to panic here
}

func CmdWaitInterrupt(cmd *cobra.Command) {
	ContextWaitInterrupt(cmd.Context())
}

func ContextWaitInterrupt(ctx context.Context) {
	contextShutter(ctx).userWaitInterrupt()
}

func CmdSoftInterrupt(cmd *cobra.Command) {
	ContextSoftInterrupt(cmd.Context())
}

func ContextSoftInterrupt(ctx context.Context) {
	contextShutter(ctx).softInterrupt()
}

func contextShutter(ctx context.Context) *shutter {
	return ctx.Value(shutterKey{}).(*shutter) //nolint:errcheck,revive // it's ok to panic here
}

func cmdInject[A App[C], C tcfg.Config](cmd *cobra.Command, app A, shut *shutter) (cancel context.CancelFunc) {
	ctx := cmd.Context()

	ctx = context.WithValue(ctx, appKey{}, app)
	ctx = context.WithValue(ctx, shutterKey{}, shut)

	ctx, cancel = context.WithCancel(ctx)
	cmd.SetContext(ctx)

	return cancel
}
