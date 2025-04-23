package config

import (
	"fmt"
)

type HTTPConfig struct {
	Address string `env:"NOTIFY_HTTP_ADDRESS"`
}

type Config struct {
	WorkersNum  int `env:"NOTIFY_WORKERS_NUM"`
	ExternalURL string `env:"NOTIFY_EXTERNAL_URL"`
	HTTPConfig  HTTPConfig
}

func (c *Config) Validate() error {
	if c.WorkersNum < 1 {
		return fmt.Errorf("WorkersNum must be greater than zero")
	}

	return nil
}
