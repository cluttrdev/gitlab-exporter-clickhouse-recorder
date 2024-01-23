package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/config"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/exporter"
	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
)

type RunConfig struct {
	RootConfig

	ServerHost string
	ServerPort string

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
}

func (c *RunConfig) Exec(ctx context.Context, args []string) error {
	log.SetOutput(c.out)

	// load configuration
	log.Printf("Loading configuration from %s\n", c.RootConfig.filename)
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
		}
	})

	writeConfig(c.out, cfg)

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

	// setup listener
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// setup server
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGitLabExporterServer(grpcServer, exporter.NewServer(client))

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
			log.Println("Got SIGINT/SIGTERM, exiting")
			cancel()
		case <-ctx.Done():
			log.Println("Done")
		}

		grpcServer.GracefulStop()
	}()

	// run
	log.Printf("Listening on %s", listener.Addr().String())
	return grpcServer.Serve(listener)

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
			log.Println("Readiness check succeeded")
			return nil
		}

		log.Println(fmt.Errorf("Readiness check failed: %w", err))
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
