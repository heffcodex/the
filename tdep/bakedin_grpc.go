package tdep

import (
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

		logDecider := grpc_zap.WithDecider(func(_ string, err error) bool { return o.IsDebug() || err != nil })
		unaryLog := grpc_zap.UnaryClientInterceptor(o.Log(), logDecider)
		streamLog := grpc_zap.StreamClientInterceptor(o.Log(), logDecider)

		dialOptions = append(dialOptions,
			grpc.WithUserAgent(o.Name()),
			grpc.WithUnaryInterceptor(unaryLog),
			grpc.WithStreamInterceptor(streamLog),
		)

		return grpc.NewClient(target, dialOptions...)
	}

	return New(resolve, options...)
}
