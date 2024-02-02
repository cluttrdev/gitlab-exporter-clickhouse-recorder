package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/config"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/exporter"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/probes"
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

	if err := waitForReady(ctx, client.CheckReadiness); err != nil {
		return err
	}

	if err := client.CreateTables(ctx); err != nil {
		return err
	}

	if err := client.InitCache(ctx); err != nil {
		return err
	}

	// setup grpc server
	grpcServer := exporter.NewServer(
		exporter.NewExporter(client),
		exporter.ServerConfig{
			Host: cfg.Server.Host,
			Port: cfg.Server.Port,
		},
	)

	// setup daemon
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		select {
		case <-signalChan:
			slog.Info("Got SIGINT/SIGTERM, exiting")
			cancel()
		case <-ctx.Done():
			slog.Info("Done")
		}
	}()

	if cfg.HTTP.Probes.Enabled {
		// setup http server
		httpServer := probes.NewServer(probes.ServerConfig{
			Host: "127.0.0.1",
			Port: "8080",

			ReadinessCheck: func() error {
				return grpcServer.CheckReadiness(ctx)
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

func waitForReady(ctx context.Context, ready func(ctx context.Context) error) error {
	var (
		maxTries         int     = 5
		backoffBaseSec   float64 = 1.0
		backoffJitterSec float64 = 1.0
	)

	ticker := time.NewTicker(time.Second)

	var err error
	for i := 0; i < maxTries; i++ {
		if err = ready(ctx); err == nil {
			slog.Debug("Readiness check succeeded")
			return nil
		}

		slog.Error("Readiness check failed", "error", err)
		delaySec := backoffBaseSec*math.Pow(2, float64(i)) + backoffJitterSec*rand.Float64()
		delay := time.Duration(delaySec) * time.Second
		ticker.Reset(delay)

		select {
		case <-ticker.C:
		case <-ctx.Done():
			ticker.Stop()
			return context.Canceled
		}
	}

	return errors.New("Failed to get ready")
}
