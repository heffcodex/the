package tdep_grpc

import (
	"strconv"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/heffcodex/the/tdep"
	"go.uber.org/zap"
)

type ClientConfig struct {
	Host string `json:"host" mapstructure:"host" yaml:"host"`
	Port uint16 `json:"port" mapstructure:"port" yaml:"port"`
}

func NewClient[C grpc.ClientConnInterface](cfg ClientConfig, dialOptions []grpc.DialOption, paramFuncs ...tdep.ParamFunc) *tdep.D[C] {
	resolve := func(_ context.Context, p tdep.Params) (C, error) {
		target := cfg.Host + ":" + strconv.FormatInt(int64(cfg.Port), 10)

		dialOptions = append(dialOptions, grpc.WithUserAgent(p.Name()))
		dialOptions = append(dialOptions, loggerDialOptions(p.Log())...)

		client, err := grpc.NewClient(target, dialOptions...)
		if err != nil {
			return *new(C), err
		}

		return any(client).(C), nil //nolint:errcheck,revive // should never panic
	}

	return tdep.New(resolve, paramFuncs...)
}

func loggerDialOptions(l *zap.Logger) []grpc.DialOption {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptorLogger(l), opts...),
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(interceptorLogger(l), opts...),
		),
	}
}

func interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
