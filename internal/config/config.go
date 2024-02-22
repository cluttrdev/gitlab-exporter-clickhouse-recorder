package config

import (
	"github.com/creasty/defaults"
)

type Config struct {
	ClickHouse ClickHouse `default:"{}" yaml:"clickhouse"`
	Server     Server     `default:"{}" yaml:"server"`
	Log        Log        `default:"{}" yaml:"log"`
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

type Log struct {
	Level  string `default:"info" yaml:"level"`
	Format string `default:"text" yaml:"format"`
}

func SetDefaults(cfg *Config) {
	defaults.MustSet(cfg)
}
