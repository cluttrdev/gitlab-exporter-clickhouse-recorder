package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/config"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/exporter"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/probes"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/retry"
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
	ctx, cancel := setupDaemon(ctx)
	// ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// load configuration
	fmt.Fprintf(c.out, "Loading configuration from %s\n", c.RootConfig.filename)
	cfg := config.Default()
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

	writeConfig(c.out, cfg)
	initLogging(c.out, cfg.Log)

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

	ErrUnspecified := errors.New("Unspecified")
	ready := readyState{}
	ready.Store(&ErrUnspecified)

	if cfg.HTTP.Probes.Enabled {
		// setup http server
		httpServer := probes.NewServer(probes.ServerConfig{
			Host: cfg.HTTP.Host,
			Port: cfg.HTTP.Port,

			ReadinessCheck: func() error {
				return *ready.Load()
			},
			LivenessCheck: func() error {
				return client.Ping(context.Background())
			},

			Debug: cfg.HTTP.Probes.Debug,
		})

		// run http server
		go func() {
			if err := httpServer.ListenAndServe(ctx); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	// getting ready
	if err := connectAndInit(ctx, client, &ready); err != nil {
		slog.Error("Failed to get ready", "error", err)
		return err
	}
	slog.Debug("Readiness check succeeded")

	// setup grpc server
	grpcServer := exporter.NewServer(
		exporter.NewExporter(client),
		exporter.ServerConfig{
			Host: cfg.Server.Host,
			Port: cfg.Server.Port,
		},
	)

	// run grpc server
	return grpcServer.ListenAndServe(ctx)
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

type readyState struct {
	atomic.Pointer[error]
}

func connectAndInit(ctx context.Context, client *clickhouse.Client, ready *readyState) error {
	// try pinging clickhouse server
	seconds := func(d time.Duration) time.Duration {
		s := math.Ceil(d.Seconds())
		return time.Duration(s) * time.Second
	}
	err := retry.Do(
		func(ctx context.Context) error {
			err := client.Ping(ctx)
			if err != nil {
				ready.Store(&err)

				args := []any{
					"error", err,
				}

				v, ok := ctx.Value(retry.ContextValuesKey("retry")).(retry.ContextValues)
				if ok {
					args = append(args, "retry.delay", seconds(v.Delay))
				}

				slog.Debug("Readiness check failed", args...)
			}
			return err
		},
		// as long as context is not done
		retry.WithContext(ctx),
		// with unlimited attempts
		retry.MaxAttempts(0),
		// with exponential backoff
		retry.WithBackoff(retry.Backoff{
			InitialDelay: 1 * time.Second,
			MaxDelay:     5 * time.Minute,
			Factor:       2.0,
			Jitter:       0.1, // +/- 10% jitter
		}),
	)
	if err != nil {
		return err
	}

	if err := client.CreateTables(ctx); err != nil {
		ready.Store(&err)
		return err
	}

	if err := client.InitCache(ctx); err != nil {
		ready.Store(&err)
		return err
	}

	ready.Store(nil)
	return nil
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
