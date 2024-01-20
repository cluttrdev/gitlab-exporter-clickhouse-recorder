package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
)

type DeduplicateConfig struct {
	RootConfig

	final       bool
	by          columnList
	except      columnList
	throwIfNoop bool
}

func NewDeduplicateCmd(out io.Writer) *cli.Command {
	fs := flag.NewFlagSet(fmt.Sprintf("%s deduplicate", exeName), flag.ContinueOnError)

	cfg := DeduplicateConfig{
		RootConfig: RootConfig{
			out: out,
		},
	}

	cfg.RegisterFlags(fs)

	return &cli.Command{
		Name:       "deduplicate",
		ShortUsage: fmt.Sprintf("%s deduplicate [option]... table", exeName),
		ShortHelp:  "Deduplicate database table",
		Flags:      fs,
		Exec:       cfg.Exec,
	}
}

func (c *DeduplicateConfig) RegisterFlags(fs *flag.FlagSet) {
	c.RootConfig.RegisterFlags(fs)

	fs.BoolVar(&c.final, "final", true, "Optimize even if all data is already in one part. (default: true)")
	fs.Var(&c.by, "by", "Comma separated list of columns to deduplicate by. (default: [])")
	fs.Var(&c.except, "except", "Comma separated list of columns to not deduplicate by. (default: [])")
	fs.BoolVar(&c.throwIfNoop, "throw-if-noop", true, "Notify if deduplication is not performed. (default: true)")
}

func (c *DeduplicateConfig) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid number of positional arguments: %v", args)
	}

	table := args[0]

	ch, err := clickhouse.NewClient(clickhouse.ClientConfig{
		Host:     c.RootConfig.Host,
		Port:     c.RootConfig.Port,
		Database: c.RootConfig.Database,
		User:     c.RootConfig.User,
		Password: c.RootConfig.Password,
	})
	if err != nil {
		return fmt.Errorf("error creating clickhouse client: %w", err)
	}

	opt := clickhouse.DeduplicateTableOptions{
		Database:    c.Database,
		Table:       table,
		Final:       &c.final,
		By:          c.by,
		Except:      c.except,
		ThrowIfNoop: &c.throwIfNoop,
	}

	return clickhouse.DeduplicateTable(ctx, opt, ch)
}

type columnList []string

func (f *columnList) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *columnList) Set(value string) error {
	values := strings.Split(value, ",")
	for _, v := range values {
		*f = append(*f, v)
	}
	return nil
}
