package the

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/heffcodex/the/tcf"
	"github.com/heffcodex/the/tdi"
	"github.com/heffcodex/the/tzap"
)

type IApp[C tcf.IConfig] interface {
	C() C
	D() *tdi.Container
	L() *zap.Logger
	Close(ctx context.Context) error
}

var _ IApp[tcf.IConfig] = (*App[tcf.IConfig])(nil)

type App[C tcf.IConfig] struct {
	cfg C
	ctn tdi.Container
	log *zap.Logger
}

func NewDefaultApp[C tcf.IConfig]() (*App[C], error) {
	var (
		configLoader = tcf.NewDefaultLoader[C]()
		logEncoder   func(cfg zapcore.EncoderConfig) zapcore.Encoder
	)

	if _, err := os.Stat("/.dockerenv"); err == nil {
		logEncoder = zapcore.NewJSONEncoder
	} else {
		logEncoder = zapcore.NewConsoleEncoder
	}

	return NewApp(configLoader, tzap.DefaultCore(logEncoder))
}

func NewApp[C tcf.IConfig](configLoader *tcf.Loader[C], zapCoreFunc tzap.CoreFunc) (*App[C], error) {
	log := zap.New(zapCoreFunc(zap.InfoLevel))
	defer func() { _ = zap.ReplaceGlobals(log) }()

	config, err := configLoader.Get()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	zapCore := zapCoreFunc(config.LogLevel())
	log = zap.New(zapCore).Named(config.AppName()).With(zap.String("env", config.AppEnv()))

	_, err = maxprocs.Set(
		maxprocs.Logger(
			func(format string, args ...any) { log.Named("maxprocs").Info(fmt.Sprintf(format, args...)) },
		),
	)
	if err != nil {
		return nil, fmt.Errorf("set maxprocs: %w", err)
	}

	return &App[C]{
		cfg: config,
		log: log,
	}, nil
}

func (a *App[C]) C() C                            { return a.cfg }
func (a *App[C]) D() *tdi.Container               { return &a.ctn }
func (a *App[C]) L() *zap.Logger                  { return a.log }
func (a *App[C]) Close(ctx context.Context) error { return a.ctn.Close(ctx) }
