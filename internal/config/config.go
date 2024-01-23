package config

import (
	"github.com/creasty/defaults"
)

type Config struct {
	ClickHouse ClickHouse `default:"{}" yaml:"clickhouse"`
	Server     Server     `default:"{}" yaml:"server"`
}

type ClickHouse struct {
	Host     string `default:"127.0.0.1" yaml:"host"`
	Port     string `default:"9000" yaml:"port"`
	Database string `default:"default" yaml:"database"`
	User     string `default:"default" yaml:"user"`
	Password string `default:"" yaml:"password"`
}

type Server struct {
	Host string `default:"0.0.0.0" yaml:"host"`
	Port string `default:"0" yaml:"port"`
}

func Default() Config {
	var cfg Config

	defaults.MustSet(&cfg)

	return cfg
}
