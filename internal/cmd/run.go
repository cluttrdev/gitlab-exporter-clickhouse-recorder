package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/cluttrdev/cli"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/config"
	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/exporter"
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

	if c.debug {
		cfg.HTTP.Enabled = true
		cfg.HTTP.Debug = true
		cfg.Log.Level = "debug"
	}

	if cfg.Log.Level == "debug" {
		writeConfig(c.out, cfg)
	}
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

	// create grpc server
	grpcServer := exporter.NewServer(
		exporter.NewExporter(client),
	)

	// setup run group
	g := &run.Group{}

	{ // serve grpc
		ctx, cancel := context.WithCancel(ctx)
		g.Add(func() error { // execute
			addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
			return grpcServer.ListenAndServe(ctx, addr)
		}, func(err error) { // interrupt
			cancel()
		})
	}

	if cfg.HTTP.Enabled { // serve http
		addr := fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
		httpServer := &http.Server{
			Addr: addr,
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(
			grpcServer.MetricsCollector(),
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)

		g.Add(func() error { // execute
			m := http.NewServeMux()

			m.Handle(
				"/metrics",
				promhttp.InstrumentMetricHandler(
					reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
				),
			)

			if cfg.HTTP.Debug {
				m.HandleFunc("/debug/pprof/", pprof.Index)
				m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
				m.HandleFunc("/debug/pprof/profile", pprof.Profile)
				m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
				m.HandleFunc("/debug/pprof/trace", pprof.Trace)
			}

			httpServer.Handler = m

			slog.Info("Starting HTTP server", "addr", httpServer.Addr)
			return httpServer.ListenAndServe()
		}, func(error) { // interrupt
			if err := httpServer.Shutdown(context.Background()); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					slog.Error(err.Error())
				}
			}
		})
	}

	{ // signal handler
		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		g.Add(func() error { // execute
			<-ctx.Done()
			return ctx.Err()
		}, func(err error) { // interrupt
			cancel()
		})
	}

	return g.Run()
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
