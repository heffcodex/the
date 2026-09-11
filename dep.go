package the

import (
	"github.com/heffcodex/the/tcfg"
	"github.com/heffcodex/the/tdep"
)

func DefaultDep[C tcfg.Config, A App[C]](app A, loggerName ...string) []tdep.ParamFunc {
	log := app.L().Named("dep")
	for _, name := range loggerName {
		log = log.Named(name)
	}

	return []tdep.ParamFunc{
		tdep.WithName(app.C().AppName()),
		tdep.WithLogger(log),
	}
}

func DefaultSingleton[C tcfg.Config, A App[C]](app A, loggerName ...string) []tdep.ParamFunc {
	return append(DefaultDep(app, loggerName...), tdep.AsSingleton())
}
