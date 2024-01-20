package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/cluttrdev/cli"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/exporter"
	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
)

type RunConfig struct {
	RootConfig

	ServerHost string
	ServerPort string
}

func NewRunCmd(out io.Writer) *cli.Command {
	cfg := RunConfig{
		RootConfig: RootConfig{
			out: out,
		},
	}

	fs := flag.NewFlagSet("run", flag.ExitOnError)
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

	// create clickhouse client
	client, err := clickhouse.NewClient(clickhouse.ClientConfig{
		Host:     c.Host,
		Port:     c.Port,
		Database: c.Database,
		User:     c.User,
		Password: c.Password,
	})
	if err != nil {
		return fmt.Errorf("error creating clickhouse client")
	}

	if err := client.CreateTables(ctx); err != nil {
		return err
	}

	if err := client.CheckReadiness(ctx); err != nil {
		return err
	}

	// setup listener
	addr := fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
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
