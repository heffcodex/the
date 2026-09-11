package tcf

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	EnvConfigPath = "CONFIG_PATH"
	EnvConfigFile = "CONFIG_FILE"
	EnvConfigType = "CONFIG_TYPE"
)

type Loader[C IConfig] struct {
	config         C
	loaded         bool
	mutex          sync.RWMutex
	viper          *viper.Viper
	decoderOptions []viper.DecoderConfigOption
}

func NewLoader[C IConfig](v *viper.Viper, decoderOptions ...viper.DecoderConfigOption) *Loader[C] {
	return &Loader[C]{viper: v, decoderOptions: decoderOptions}
}

func NewDefaultLoader[C IConfig]() *Loader[C] {
	v := viper.New()

	if configFile, ok := os.LookupEnv(EnvConfigFile); ok { //nolint:nestif // ok
		v.SetConfigFile(configFile)
	} else {
		if configPath, ok := os.LookupEnv(EnvConfigPath); ok {
			v.AddConfigPath(configPath)
		} else {
			v.AddConfigPath(".")
			v.AddConfigPath("./.data")
			v.AddConfigPath("./.config.d")
			v.AddConfigPath("./.mnt/config.d")
		}

		if configType, ok := os.LookupEnv(EnvConfigType); ok {
			v.SetConfigType(configType)
		} else {
			v.SetConfigType("yaml")
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return NewLoader[C](v)
}

func (l *Loader[C]) Must() C {
	c, err := l.Get()
	if err != nil {
		panic(err)
	}

	return c
}

func (l *Loader[C]) Get() (C, error) {
	l.mutex.RLock()

	if l.loaded {
		defer l.mutex.RUnlock()
		return l.config, nil
	}

	l.mutex.RUnlock()
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.loaded {
		return l.config, nil
	}

	if err := l.load(); err != nil {
		return *new(C), fmt.Errorf("load config: %w", err)
	}

	return l.config, nil
}

func (l *Loader[C]) load() error {
	var config C

	if err := config.BeforeRead(l.viper); err != nil {
		return fmt.Errorf("before read: %w", err)
	}

	if err := l.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	decoderOptions := []viper.DecoderConfigOption{ //nolint:prealloc // @formatter:off // ok
		viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()),
	}

	if err := l.viper.Unmarshal(&config, append(decoderOptions, l.decoderOptions...)...); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if err := config.AfterRead(l.viper); err != nil {
		return fmt.Errorf("after read: %w", err)
	}

	l.config = config
	l.loaded = true

	return nil
}
