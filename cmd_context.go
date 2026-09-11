package the

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/heffcodex/the/tcf"
)

type (
	appKey     struct{}
	shutterKey struct{}
)

func CmdApp[A IApp[C], C tcf.IConfig](cmd *cobra.Command) A {
	return CtxApp[A, C](cmd.Context())
}

func CtxApp[A IApp[C], C tcf.IConfig](ctx context.Context) A {
	return ctx.Value(appKey{}).(A) //nolint:errcheck,revive // it's ok to panic here
}

func CmdWaitInterrupt(cmd *cobra.Command) {
	CtxWaitInterrupt(cmd.Context())
}

func CtxWaitInterrupt(ctx context.Context) {
	ctxShutter(ctx).userWaitInterrupt()
}

func CmdSoftInterrupt(cmd *cobra.Command) {
	CtxSoftInterrupt(cmd.Context())
}

func CtxSoftInterrupt(ctx context.Context) {
	ctxShutter(ctx).softInterrupt()
}

func ctxShutter(ctx context.Context) *shutter {
	return ctx.Value(shutterKey{}).(*shutter) //nolint:errcheck,revive // it's ok to panic here
}

func cmdInject[A IApp[C], C tcf.IConfig](cmd *cobra.Command, app A, shut *shutter) (cancel context.CancelFunc) {
	ctx := cmd.Context()

	ctx = context.WithValue(ctx, appKey{}, app)
	ctx = context.WithValue(ctx, shutterKey{}, shut)

	ctx, cancel = context.WithCancel(ctx)
	cmd.SetContext(ctx)

	return cancel
}
