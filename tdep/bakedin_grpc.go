package tdep

import (
	"context"
	"strconv"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
)

type GRPCConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
	Port uint16 `mapstructure:"port" json:"port" yaml:"port"`
}

func NewGRPC(cfg GRPCConfig, dialOptions []grpc.DialOption, options ...Option) *D[*grpc.ClientConn] {
	resolve := func(o OptSet) (*grpc.ClientConn, error) {
		target := cfg.Host + ":" + strconv.FormatInt(int64(cfg.Port), 10)

		var (
			unaryLog  grpc.UnaryClientInterceptor
			streamLog grpc.StreamClientInterceptor
		)

		if o.IsDebug() {
			decider := func(context.Context, string) bool { return true }

			unaryLog = grpc_zap.PayloadUnaryClientInterceptor(o.Log(), decider)
			streamLog = grpc_zap.PayloadStreamClientInterceptor(o.Log(), decider)
		} else {
			optDecider := grpc_zap.WithDecider(func(_ string, err error) bool { return err != nil })

			unaryLog = grpc_zap.UnaryClientInterceptor(o.Log(), optDecider)
			streamLog = grpc_zap.StreamClientInterceptor(o.Log(), optDecider)
		}

		dialOptions = append(dialOptions,
			grpc.WithUserAgent(o.Name()),
			grpc.WithUnaryInterceptor(unaryLog),
			grpc.WithStreamInterceptor(streamLog),
		)

		return grpc.NewClient(target, dialOptions...)
	}

	return New(resolve, options...)
}
