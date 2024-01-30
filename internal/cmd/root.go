package cmd

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"

	"github.com/cluttrdev/cli"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/config"
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
	Host     string
	Port     string
	Database string
	User     string
	Password string

	filename string
	out      io.Writer
	flags    *flag.FlagSet
}

func (c *RootConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Host, "clickhouse-host", "127.0.0.1", "The ClickHouse server name (default: '127.0.0.1').")
	fs.StringVar(&c.Port, "clickhouse-port", "9000", "The ClickHouse port to connect to (default: '9000')")
	fs.StringVar(&c.Database, "clickhouse-database", "default", "Select the current default ClickHouse database (default: 'default').")
	fs.StringVar(&c.User, "clickhouse-user", "default", "The ClickHouse username to connect with (default: 'default').")
	fs.StringVar(&c.Password, "clickhouse-password", "", "The ClickHouse password (default: '').")

	fs.StringVar(&c.filename, "config", "", "The configuration file to use.")
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
	fmt.Fprintln(out, "----")
	fmt.Fprintf(out, "ClickHouse Host: %s\n", cfg.ClickHouse.Host)
	fmt.Fprintf(out, "ClickHouse Port: %s\n", cfg.ClickHouse.Port)
	fmt.Fprintf(out, "ClickHouse Database: %s\n", cfg.ClickHouse.Database)
	fmt.Fprintf(out, "ClickHouse User: %s\n", cfg.ClickHouse.User)
	fmt.Fprintf(out, "ClickHouse Password: %x\n", sha256String(cfg.ClickHouse.Password))
	fmt.Fprintln(out, "----")
	fmt.Fprintf(out, "Server Host: %s\n", cfg.Server.Host)
	fmt.Fprintf(out, "Server Port: %s\n", cfg.Server.Port)
	fmt.Fprintln(out, "----")
	fmt.Fprintf(out, "Log Level: %s\n", cfg.Log.Level)
	fmt.Fprintf(out, "Log Format: %s\n", cfg.Log.Format)
	fmt.Fprintln(out, "----")
}

func sha256String(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}
