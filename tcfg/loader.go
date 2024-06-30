package tcfg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Loader[C Config] struct {
	config C
	loaded bool
	mutex  sync.RWMutex
	viper  *viper.Viper
}

func NewLoader[C Config](v *viper.Viper) *Loader[C] {
	return &Loader[C]{viper: v}
}

func NewDefaultLoader[C Config]() *Loader[C] {
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("./.data")
	v.AddConfigPath("./.mnt/config.d")
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvPrefix("CFG")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return NewLoader[C](v)
}

func (l *Loader[C]) LoadOnce() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.loaded {
		return nil
	}

	return l.load()
}

func (l *Loader[C]) load() error {
	var config C

	if err := config.BeforeRead(l.viper); err != nil {
		return fmt.Errorf("before read: %w", err)
	}

	if err := l.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	if err := l.viper.UnmarshalExact(&config); err != nil {
		return fmt.Errorf("unmarshal exact: %w", err)
	}

	if err := config.AfterRead(l.viper); err != nil {
		return fmt.Errorf("after read: %w", err)
	}

	l.config = config
	l.loaded = true

	return nil
}

func (l *Loader[C]) Get() C {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if !l.loaded {
		panic("config not loaded")
	}

	return l.config
}
