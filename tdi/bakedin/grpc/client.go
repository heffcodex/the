package tdi_grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/heffcodex/the/tdi"
)

type ClientConfig struct {
	Host string `json:"host" mapstructure:"host" yaml:"host"`
	Port uint16 `json:"port" mapstructure:"port" yaml:"port"`
}

func NewClient[C grpc.ClientConnInterface](cfg ClientConfig, dialOptions []grpc.DialOption, paramFuncs ...tdi.ParamFunc) *tdi.D[C] {
	resolve := func(_ context.Context, p tdi.Params) (C, error) {
		target := cfg.Host + ":" + strconv.FormatUint(uint64(cfg.Port), 10)

		dialOptions = append(dialOptions, grpc.WithUserAgent(p.Name()))
		dialOptions = append(dialOptions, loggerDialOptions(p.Log())...)

		client, err := grpc.NewClient(target, dialOptions...)
		if err != nil {
			return *new(C), err
		}

		return any(client).(C), nil //nolint:errcheck,revive // should never panic
	}

	return tdi.NewWithHealthCheck(resolve, func(context.Context, *tdi.D[C]) error {
		return nil // TODO
	}, paramFuncs...)
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
			key := fields[i].(string) //nolint: revive // should never panic
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key, v))
			case int:
				f = append(f, zap.Int(key, v))
			case bool:
				f = append(f, zap.Bool(key, v))
			default:
				f = append(f, zap.Any(key, v))
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
