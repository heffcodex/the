package the

import (
	"github.com/heffcodex/the/tcfg"
	"github.com/heffcodex/the/tdep"
)

func DefaultDepOptions[C tcfg.Config, A App[C]](app A, loggerName ...string) []tdep.Option {
	log := app.L().Named("dep")
	for _, name := range loggerName {
		log = log.Named(name)
	}

	return []tdep.Option{
		tdep.Name(app.C().AppName()),
		tdep.Env(app.C().AppEnv()),
		tdep.Log(log),
	}
}

func DefaultDepSingleton[C tcfg.Config, A App[C]](app A, loggerName ...string) []tdep.Option {
	return append(DefaultDepOptions(app, loggerName...), tdep.Singleton())
}
