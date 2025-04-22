package tcfg

type App struct {
	Name            string `mapstructure:"name" json:"name" yaml:"name"`
	Key             Key    `mapstructure:"key" json:"key" yaml:"key"`
	Env             Env    `mapstructure:"env" json:"env" yaml:"env"`
	LogLevel        string `mapstructure:"logLevel" json:"logLevel" yaml:"logLevel"`
	ShutdownTimeout int    `mapstructure:"shutdownTimeout" json:"shutdownTimeout" yaml:"shutdownTimeout"`
}
