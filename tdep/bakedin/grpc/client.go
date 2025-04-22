package tdep_grpc

import (
	"strconv"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"

	"github.com/heffcodex/the/tdep"
)

type ClientConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
	Port uint16 `mapstructure:"port" json:"port" yaml:"port"`
}

func NewClient[C grpc.ClientConnInterface](cfg ClientConfig, dialOptions []grpc.DialOption, options ...tdep.Option) *tdep.D[C] {
	resolve := func(o tdep.OptSet) (C, error) {
		target := cfg.Host + ":" + strconv.FormatInt(int64(cfg.Port), 10)

		logDecider := grpc_zap.WithDecider(func(_ string, err error) bool { return o.IsDebug() || err != nil })
		unaryLog := grpc_zap.UnaryClientInterceptor(o.Log(), logDecider)
		streamLog := grpc_zap.StreamClientInterceptor(o.Log(), logDecider)

		dialOptions = append(dialOptions,
			grpc.WithUserAgent(o.Name()),
			grpc.WithUnaryInterceptor(unaryLog),
			grpc.WithStreamInterceptor(streamLog),
		)

		client, err := grpc.NewClient(target, dialOptions...)
		if err != nil {
			return *new(C), err
		}

		return any(client).(C), nil //nolint:errcheck,revive // should never panic
	}

	return tdep.New(resolve, options...)
}
