package tdi_bun

import "time"

type Config struct {
	DSN            string `json:"dsn"            mapstructure:"dsn"            yaml:"dsn"`
	MaxConnections int    `json:"maxConnections" mapstructure:"maxConnections" yaml:"maxConnections"`
	MaxIdleTime    int    `json:"maxIdleTime"    mapstructure:"maxIdleTime"    yaml:"maxIdleTime"`
}

func (c *Config) MaxIdleTimeSeconds() time.Duration {
	if c.MaxIdleTime < 1 {
		return 0
	}

	return time.Duration(c.MaxIdleTime) * time.Second
}
