package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/cluttrdev/cli"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/config"
)

func NewRootCmd(out io.Writer) *cli.Command {
	fs := flag.NewFlagSet(exeName, flag.ExitOnError)

	cfg := RootConfig{
		out:   out,
		flags: fs,
	}
	cfg.RegisterFlags(fs)

	return &cli.Command{
		Name:  exeName,
		Flags: fs,
		Exec:  cfg.Exec,
	}
}

type RootConfig struct {
	filename string
	out      io.Writer
	flags    *flag.FlagSet
	debug    bool
}

func (c *RootConfig) RegisterFlags(fs *flag.FlagSet) {
	_ = fs.String("clickhouse-host", "127.0.0.1", "The ClickHouse server name (default: '127.0.0.1').")
	_ = fs.String("clickhouse-port", "9000", "The ClickHouse port to connect to (default: '9000')")
	_ = fs.String("clickhouse-database", "default", "Select the current default ClickHouse database (default: 'default').")
	_ = fs.String("clickhouse-user", "default", "The ClickHouse username to connect with (default: 'default').")
	_ = fs.String("clickhouse-password", "", "The ClickHouse password (default: '').")

	fs.StringVar(&c.filename, "config", "", "The configuration file to use.")

	fs.BoolVar(&c.debug, "debug", false, "Run in debug mode.")
}

func (c *RootConfig) Exec(ctx context.Context, args []string) error {
	return flag.ErrHelp
}

func loadConfig(filename string, flags *flag.FlagSet, cfg *config.Config) error {
	// load configuration file first
	if filename != "" {
		if err := config.LoadFile(filename, cfg); err != nil {
			return err
		}
	}

	// override with values passed as env vars or flags
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "clickhouse-host":
			cfg.ClickHouse.Host = f.Value.String()
		case "clickhouse-port":
			cfg.ClickHouse.Port = f.Value.String()
		case "clickhouse-database":
			cfg.ClickHouse.Database = f.Value.String()
		case "clickhouse-user":
			cfg.ClickHouse.User = f.Value.String()
		case "clickhouse-password":
			cfg.ClickHouse.Password = f.Value.String()
		}
	})

	return nil
}

func writeConfig(out io.Writer, cfg config.Config) {
	_cfg := cfg
	_cfg.ClickHouse.Password = fmt.Sprintf("%x", sha256String(cfg.ClickHouse.Password))

	b, err := json.MarshalIndent(_cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(out, "error marshalling config: %v\n", err)
	}
	fmt.Fprint(out, string(b))
}

func sha256String(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}
