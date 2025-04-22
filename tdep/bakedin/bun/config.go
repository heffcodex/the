package tdep_bun

import "time"

type Config struct {
	DSN            string `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	MaxConnections int    `mapstructure:"maxConnections" json:"maxConnections" yaml:"maxConnections"`
	MaxIdleTime    int    `mapstructure:"maxIdleTime" json:"maxIdleTime" yaml:"maxIdleTime"`
}

func (c *Config) MaxIdleTimeSeconds() time.Duration {
	if c.MaxIdleTime < 1 {
		return 0
	}

	return time.Duration(c.MaxIdleTime) * time.Second
}
