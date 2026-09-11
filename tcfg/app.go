package tcfg

type App struct {
	Name            string `json:"name"            mapstructure:"name"            yaml:"name"`
	Key             Key    `json:"key"             mapstructure:"key"             yaml:"key"`
	Env             Env    `json:"env"             mapstructure:"env"             yaml:"env"`
	LogLevel        string `json:"logLevel"        mapstructure:"logLevel"        yaml:"logLevel"`
	ShutdownTimeout int    `json:"shutdownTimeout" mapstructure:"shutdownTimeout" yaml:"shutdownTimeout"`
}
