package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cluttrdev/cli"
)

const exeName string = "gitlab-clickhouse-exporter"

func Execute() error {
	out := os.Stderr

	root := NewRootCmd(out)
	root.Subcommands = []*cli.Command{
		NewRunCmd(out),
		NewDeduplicateCmd(out),
	}

	args := os.Args[1:]
	opts := []cli.ParseOption{}

	if err := root.Parse(args, opts...); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		} else {
			return fmt.Errorf("error parsing arguments: %w", err)
		}
	}

	return root.Run(context.Background())
}

func NewRootCmd(out io.Writer) *cli.Command {
	cfg := RootConfig{}

	fs := flag.NewFlagSet(exeName, flag.ExitOnError)
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

	out io.Writer
}

func (c *RootConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Host, "host", "127.0.0.1", "The ClickHouse server name (default: '127.0.0.1').")
	fs.StringVar(&c.Port, "port", "9000", "The ClickHouse port to connect to (default: '9000')")
	fs.StringVar(&c.Database, "database", "default", "Select the current default ClickHouse database (default: 'default').")
	fs.StringVar(&c.User, "user", "default", "The ClickHouse username to connect with (default: 'default').")
	fs.StringVar(&c.Password, "password", "", "The ClickHouse password (default: '').")
}

func (c *RootConfig) Exec(ctx context.Context, args []string) error {
	return flag.ErrHelp
}
