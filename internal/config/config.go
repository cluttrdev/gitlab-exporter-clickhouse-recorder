package config

import (
	"github.com/creasty/defaults"
)

type Config struct {
	ClickHouse ClickHouse `default:"{}" yaml:"clickhouse"`
	Server     Server     `default:"{}" yaml:"server"`
	Log        Log        `default:"{}" yaml:"log"`
	HTTP       HTTP       `default:"{}" yaml:"http"`
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

type HTTP struct {
	Host   string     `default:"0.0.0.0" yaml:"host"`
	Port   string     `default:"8080" yaml:"port"`
	Probes HTTPProbes `default:"{}" yaml:"probes"`
}

type HTTPProbes struct {
	Enabled bool `default:"true" yaml:"enabled"`
	Debug   bool `default:"false" yaml:"debug"`
}

func Default() Config {
	var cfg Config

	defaults.MustSet(&cfg)

	return cfg
}
