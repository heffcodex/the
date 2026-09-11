package tcfg

import "go.uber.org/zap/zapcore"

type App struct {
	Name            string        `json:"name"            mapstructure:"name"            yaml:"name"`
	Env             string        `json:"env"             mapstructure:"env"             yaml:"env"`
	Key             Key           `json:"key"             mapstructure:"key"             yaml:"key"`
	LogLevel        zapcore.Level `json:"logLevel"        mapstructure:"logLevel"        yaml:"logLevel"`
	ShutdownTimeout int           `json:"shutdownTimeout" mapstructure:"shutdownTimeout" yaml:"shutdownTimeout"`
}
