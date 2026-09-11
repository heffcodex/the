package tcf

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type IConfig interface {
	AppName() string
	AppEnv() string
	AppKey() Key
	LogLevel() zapcore.Level
	ShutdownTimeout() time.Duration

	BeforeRead(v *viper.Viper) error
	AfterRead(v *viper.Viper) error
}

var _ IConfig = (*Config)(nil)

type App struct {
	Name            string        `json:"name"            mapstructure:"name"            yaml:"name"`
	Env             string        `json:"env"             mapstructure:"env"             yaml:"env"`
	Key             Key           `json:"key"             mapstructure:"key"             yaml:"key"`
	LogLevel        zapcore.Level `json:"logLevel"        mapstructure:"logLevel"        yaml:"logLevel"`
	ShutdownTimeout int           `json:"shutdownTimeout" mapstructure:"shutdownTimeout" yaml:"shutdownTimeout"`
}

type Config struct {
	App App `json:"app" mapstructure:"app" yaml:"app"`
}

func (c Config) AppName() string {
	if c.App.Name == "" {
		return filepath.Base(os.Args[0])
	}

	return c.App.Name
}

func (c Config) AppEnv() string {
	if c.App.Env == "" {
		return "dev"
	}

	return c.App.Env
}

func (c Config) AppKey() Key {
	return c.App.Key
}

func (c Config) LogLevel() zapcore.Level {
	return c.App.LogLevel
}

const appShutdownTimeoutDefault = 30 * time.Second

func (c Config) ShutdownTimeout() time.Duration {
	if c.App.ShutdownTimeout < 1 {
		return appShutdownTimeoutDefault
	}

	return time.Duration(c.App.ShutdownTimeout) * time.Second
}

func (Config) BeforeRead(_ *viper.Viper) error {
	return nil
}

func (c Config) AfterRead(_ *viper.Viper) error {
	if err := c.AppKey().Validate(); err != nil {
		return fmt.Errorf("validate app key: %w", err)
	}

	return nil
}
