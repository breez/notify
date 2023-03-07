package config

type Config struct {
	WorkersNum int `env:"NOTIFY_WORKERS_NUM"`
}
