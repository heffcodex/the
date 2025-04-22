package tcfg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config interface {
	AppName() string
	AppKey() Key
	AppEnv() Env
	LogLevel() string
	ShutdownTimeout() time.Duration

	BeforeRead(v *viper.Viper) error
	AfterRead(v *viper.Viper) error
}

const (
	AppEnvDefault             = EnvDev
	AppLogLevelDefault        = zap.InfoLevel
	AppShutdownTimeoutDefault = 30 * time.Second
)

var _ Config = (*BaseConfig)(nil)

type BaseConfig struct {
	App App `mapstructure:"app" json:"app" yaml:"app"`
}

func (c BaseConfig) AppName() string {
	if c.App.Name == "" {
		return filepath.Base(os.Args[0])
	}

	return c.App.Name
}

func (c BaseConfig) AppKey() Key {
	return c.App.Key
}

func (c BaseConfig) AppEnv() Env {
	if c.App.Env == "" {
		return AppEnvDefault
	}

	return c.App.Env
}

func (c BaseConfig) LogLevel() string {
	if c.App.LogLevel == "" {
		return AppLogLevelDefault.String()
	}

	return c.App.LogLevel
}

func (c BaseConfig) ShutdownTimeout() time.Duration {
	if c.App.ShutdownTimeout < 1 {
		return AppShutdownTimeoutDefault
	}

	return time.Duration(c.App.ShutdownTimeout) * time.Second
}

func (BaseConfig) BeforeRead(*viper.Viper) error {
	return nil
}

func (c BaseConfig) AfterRead(*viper.Viper) error {
	if err := c.AppKey().Validate(); err != nil {
		return fmt.Errorf("validate app key: %w", err)
	}

	return nil
}
