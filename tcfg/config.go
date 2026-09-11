package tcfg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type Config interface {
	AppName() string
	AppEnv() string
	AppKey() Key
	LogLevel() zapcore.Level
	ShutdownTimeout() time.Duration

	BeforeRead(v *viper.Viper) error
	AfterRead(v *viper.Viper) error
}

var _ Config = (*BaseConfig)(nil)

type BaseConfig struct {
	App App `json:"app" mapstructure:"app" yaml:"app"`
}

func (c BaseConfig) AppName() string {
	if c.App.Name == "" {
		return filepath.Base(os.Args[0])
	}

	return c.App.Name
}

func (c BaseConfig) AppEnv() string {
	if c.App.Env == "" {
		return "dev"
	}

	return c.App.Env
}

func (c BaseConfig) AppKey() Key {
	return c.App.Key
}

func (c BaseConfig) LogLevel() zapcore.Level {
	return c.App.LogLevel
}

const appShutdownTimeoutDefault = 15 * time.Second

func (c BaseConfig) ShutdownTimeout() time.Duration {
	if c.App.ShutdownTimeout < 1 {
		return appShutdownTimeoutDefault
	}

	return time.Duration(c.App.ShutdownTimeout) * time.Second
}

func (BaseConfig) BeforeRead(_ *viper.Viper) error {
	return nil
}

func (c BaseConfig) AfterRead(_ *viper.Viper) error {
	if err := c.AppKey().Validate(); err != nil {
		return fmt.Errorf("validate app key: %w", err)
	}

	return nil
}
