package the

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/heffcodex/the/tcfg"
	"github.com/heffcodex/the/tzap"
)

var (
	ErrClosed = errors.New("app is already closed")
)

type CloseFunc func(context.Context) error

type App[C tcfg.Config] interface {
	C() C
	L() *zap.Logger
	AddCloser(fns ...CloseFunc)
	Close(ctx context.Context) error
}

var _ App[tcfg.Config] = (*BaseApp[tcfg.Config])(nil)

type BaseApp[C tcfg.Config] struct {
	cfg C
	log *zap.Logger

	closed   bool
	closers  []CloseFunc
	closerMu sync.Mutex
}

func NewBaseApp[C tcfg.Config](loader *tcfg.Loader[C]) (*BaseApp[C], error) {
	if err := loader.LoadOnce(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	config := loader.Get()

	logLevel, err := zap.ParseAtomicLevel(config.LogLevel())
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var (
		appEnv  = config.AppEnv()
		zapCfg  = tzap.DefaultStdCoreConfig(logLevel)
		zapCore zapcore.Core
	)

	switch appEnv {
	case tcfg.EnvProd, tcfg.EnvStage, tcfg.EnvTest:
		zapCore = zapCfg.JSON()
	case tcfg.EnvDev:
		fallthrough
	default:
		zapCore = zapCfg.Console()
	}

	log := zap.New(zapCore).Named(config.AppName()).With(zap.String("env", appEnv.String()))
	zap.ReplaceGlobals(log)

	_, err = maxprocs.Set(
		maxprocs.Logger(
			func(format string, args ...any) { log.Named("maxprocs").Info(fmt.Sprintf(format, args...)) },
		),
	)
	if err != nil {
		return nil, fmt.Errorf("set maxprocs: %w", err)
	}

	return &BaseApp[C]{
		cfg: config,
		log: log,
	}, nil
}

func (a *BaseApp[C]) C() C           { return a.cfg }
func (a *BaseApp[C]) L() *zap.Logger { return a.log }

func (a *BaseApp[C]) AddCloser(fns ...CloseFunc) {
	_ = a.closerSafe(func() error {
		a.closers = append(a.closers, fns...)
		return nil
	})
}

func (a *BaseApp[C]) Close(ctx context.Context) error {
	return a.closerSafe(func() error {
		errs := make([]error, 0, len(a.closers))

		for i := len(a.closers) - 1; i >= 0; i-- {
			closer := a.closers[i]

			if err := closer(ctx); err != nil {
				errs = append(errs, fmt.Errorf("%d: %w", i, err))
			}
		}

		a.closed = true

		return errors.Join(errs...)
	})
}

func (a *BaseApp[C]) closerSafe(f func() error) error {
	a.closerMu.Lock()
	defer a.closerMu.Unlock()

	if a.closed {
		return ErrClosed
	}

	return f()
}
