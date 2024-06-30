package tcfg

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
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

var _ Config = (*BaseConfig)(nil)

type BaseConfig struct {
	App App `mapstructure:"app" json:"app" yaml:"app"`
}

func (c BaseConfig) AppName() string {
	return c.App.Name
}

func (c BaseConfig) AppKey() Key {
	return c.App.Key
}

func (c BaseConfig) AppEnv() Env {
	return c.App.Env
}

func (c BaseConfig) LogLevel() string {
	return c.App.LogLevel
}

func (c BaseConfig) ShutdownTimeout() time.Duration {
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
