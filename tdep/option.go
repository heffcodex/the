package tdep

import (
	"go.uber.org/zap"
)

type Params struct {
	name      string
	singleton bool
	log       *zap.Logger
}

func newParams(paramFuncs ...ParamFunc) Params {
	var params Params

	for _, fn := range paramFuncs {
		fn(&params)
	}

	if params.log == nil {
		params.log = zap.L()
	}

	return params
}

func (o Params) Name() string      { return o.name }
func (o Params) IsSingleton() bool { return o.singleton }
func (o Params) Log() *zap.Logger  { return o.log }
func (o Params) IsDebug() bool     { return o.log.Core().Enabled(zap.DebugLevel) }

type ParamFunc func(*Params)

func WithName(name string) ParamFunc {
	return func(o *Params) {
		o.name = name
	}
}

func AsSingleton() ParamFunc {
	return func(o *Params) {
		o.singleton = true
	}
}

func WithLogger(log *zap.Logger) ParamFunc {
	return func(o *Params) {
		o.log = log
	}
}
