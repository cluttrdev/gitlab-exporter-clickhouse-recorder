package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/config"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/exporter"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/probes"
)

type RunConfig struct {
	RootConfig

	ServerHost string
	ServerPort string

	LogLevel  string
	LogFormat string

	flags *flag.FlagSet
}

func NewRunCmd(out io.Writer) *cli.Command {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	cfg := RunConfig{
		RootConfig: RootConfig{
			out: out,
		},

		flags: fs,
	}
	cfg.RegisterFlags(fs)

	return &cli.Command{
		Name:       "run",
		ShortUsage: fmt.Sprintf("%s run [option]...", exeName),
		ShortHelp:  "Run gRPC server",
		Flags:      fs,
		Exec:       cfg.Exec,
	}
}

func (c *RunConfig) RegisterFlags(fs *flag.FlagSet) {
	c.RootConfig.RegisterFlags(fs)

	fs.StringVar(&c.ServerHost, "server-host", "127.0.0.1", "The gRPC server listen host. (default: '127.0.0.1')")
	fs.StringVar(&c.ServerPort, "server-port", "0", "The gRPC server listen port. (default: '0', random)")

	fs.StringVar(&c.LogLevel, "log-level", "info", "The logging level, one of 'debug', 'info', 'warn', 'error'. (default: 'info')")
	fs.StringVar(&c.LogFormat, "log-format", "text", "The logging format, either 'text' or 'json'. (default: 'text')")
}

func (c *RunConfig) Exec(ctx context.Context, args []string) error {
	// setup daemon
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// load configuration
	var cfg config.Config
	config.SetDefaults(&cfg)
	if err := loadConfig(c.RootConfig.filename, c.flags, &cfg); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}
	// override values passed as env vars or flags
	c.flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server-host":
			cfg.Server.Host = f.Value.String()
		case "server-port":
			cfg.Server.Port = f.Value.String()
		case "log-level":
			cfg.Log.Level = f.Value.String()
		case "log-format":
			cfg.Log.Format = f.Value.String()
		}
	})

	if cfg.Log.Level == "debug" {
		writeConfig(c.out, cfg)
	}
	initLogging(c.out, cfg.Log)

	if cfg.HTTP.Enabled {
		go func() {
			if err := serveHTTP(ctx, cfg.HTTP); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	// create clickhouse client
	client, err := clickhouse.NewClient(clickhouse.ClientConfig{
		Host:     cfg.ClickHouse.Host,
		Port:     cfg.ClickHouse.Port,
		Database: cfg.ClickHouse.Database,
		User:     cfg.ClickHouse.User,
		Password: cfg.ClickHouse.Password,
	})
	if err != nil {
		return fmt.Errorf("error creating clickhouse client")
	}

	// setup grpc server
	grpcServer := exporter.NewServer(
		exporter.NewExporter(client),
	)

	// run grpc server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	return grpcServer.ListenAndServe(ctx, addr)
}

func initLogging(out io.Writer, cfg config.Log) {
	if out == nil {
		out = os.Stderr
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch cfg.Format {
	case "text":
		handler = slog.NewTextHandler(out, &opts)
	case "json":
		handler = slog.NewJSONHandler(out, &opts)
	default:
		handler = slog.NewTextHandler(out, &opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func setupDaemon(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-signalChan:
			slog.Debug("Got SIGINT/SIGTERM, exiting")
			signal.Stop(signalChan)
			cancel()
		case <-ctx.Done():
			slog.Debug("Done")
		}
	}()

	return ctx, cancel
}

func serveHTTP(ctx context.Context, cfg config.HTTP) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := probes.NewServer(probes.ServerConfig{
		Address: addr,
		Debug:   cfg.Debug,
	})
	return srv.ListenAndServe(ctx)
}
