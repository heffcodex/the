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
	"github.com/heffcodex/the/tdep"
	"github.com/heffcodex/the/tzap"
)

var (
	ErrClosed = errors.New("app is already closed")
)

type (
	CloseFunc func(context.Context) error
)

type App[C tcfg.Config] interface {
	tdep.C

	C() C
	L() *zap.Logger
	AddCloser(fns ...CloseFunc)
	Close(ctx context.Context) error
}

var _ App[tcfg.Config] = (*BaseApp[tcfg.Config])(nil)

type BaseApp[C tcfg.Config] struct {
	tdep.Container

	cfg C
	log *zap.Logger

	closed   bool
	closers  []CloseFunc
	closerMu sync.RWMutex
}

func NewBaseApp[C tcfg.Config](configLoader *tcfg.Loader[C]) (*BaseApp[C], error) {
	log := zap.New(tzap.DefaultStdCoreConfig(zap.InfoLevel).Console())
	defer zap.ReplaceGlobals(log)

	config, err := configLoader.Get()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logLevel, err := zap.ParseAtomicLevel(config.LogLevel())
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var (
		appEnv  = config.AppEnv()
		zapCfg  = tzap.DefaultStdCoreConfig(logLevel)
		zapCore zapcore.Core
	)

	if appEnv == tcfg.EnvDev {
		zapCore = zapCfg.Console()
	} else {
		zapCore = zapCfg.JSON()
	}

	log = zap.New(zapCore).Named(config.AppName()).With(zap.String("env", appEnv.String()))

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
	a.closerMu.Lock()
	defer a.closerMu.Unlock()

	if !a.closed {
		a.closers = append(a.closers, fns...)
	}
}

func (a *BaseApp[C]) Close(ctx context.Context) error {
	a.closerMu.Lock()
	defer a.closerMu.Unlock()

	if a.closed {
		return ErrClosed
	}

	var errs error

	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i](ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("app[%d]: %w", i, err))
		}
	}

	if err := a.Container.Close(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("ctn: %w", err))
	}

	a.closed = true

	return errs
}
