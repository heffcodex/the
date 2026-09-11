package the

import (
	"fmt"
	"github.com/heffcodex/the/tcfg"
	"github.com/heffcodex/the/tdep"
	"github.com/heffcodex/the/tzap"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

type App[C tcfg.Config] interface {
	C() C
	D() *tdep.Container
	L() *zap.Logger
}

var _ App[tcfg.Config] = (*BaseApp[tcfg.Config])(nil)

type BaseApp[C tcfg.Config] struct {
	cfg C
	ctn tdep.Container
	log *zap.Logger
}

func NewBaseApp[C tcfg.Config](configLoader *tcfg.Loader[C], zapCoreFunc tzap.CoreFunc) (*BaseApp[C], error) {
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

	return &BaseApp[C]{
		cfg: config,
		log: log,
	}, nil
}

func (a *BaseApp[C]) C() C               { return a.cfg }
func (a *BaseApp[C]) D() *tdep.Container { return &a.ctn }
func (a *BaseApp[C]) L() *zap.Logger     { return a.log }
