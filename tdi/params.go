package tdi

import (
	"strings"

	"go.uber.org/zap"
)

type Params struct {
	name string
	log  *zap.Logger
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

func (o Params) Name() string     { return o.name }
func (o Params) Log() *zap.Logger { return o.log }
func (o Params) IsDebug() bool    { return o.log.Core().Enabled(zap.DebugLevel) }

type ParamFunc func(*Params)

func WithName(name string) ParamFunc {
	return func(o *Params) {
		o.name = name
	}
}

func WithLogger(log *zap.Logger) ParamFunc {
	return func(o *Params) {
		o.log = log
	}
}

func DefaultParams(l *zap.Logger, name ...string) []ParamFunc {
	log := l.Named("dep")
	for _, n := range name {
		log = log.Named(n)
	}

	return []ParamFunc{
		WithName(strings.Join(name, ".")),
		WithLogger(log),
	}
}
