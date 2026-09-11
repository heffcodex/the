package tdep_grpc

import (
	"strconv"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"

	"github.com/heffcodex/the/tdep"
)

type ClientConfig struct {
	Host string `json:"host" mapstructure:"host" yaml:"host"`
	Port uint16 `json:"port" mapstructure:"port" yaml:"port"`
}

func NewClient[C grpc.ClientConnInterface](cfg ClientConfig, dialOptions []grpc.DialOption, options ...tdep.ParamFunc) *tdep.D[C] {
	resolve := func(p tdep.Params) (C, error) {
		target := cfg.Host + ":" + strconv.FormatInt(int64(cfg.Port), 10)

		logDecider := grpc_zap.WithDecider(func(_ string, err error) bool { return p.IsDebug() || err != nil })
		unaryLog := grpc_zap.UnaryClientInterceptor(p.Log(), logDecider)
		streamLog := grpc_zap.StreamClientInterceptor(p.Log(), logDecider)

		dialOptions = append(dialOptions,
			grpc.WithUserAgent(p.Name()),
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
