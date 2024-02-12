package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
)

type ServerConfig struct {
	Host string
	Port string
}

type Server struct {
	config   ServerConfig
	exporter *ClickHouseExporter
}

func NewServer(exp *ClickHouseExporter, cfg ServerConfig) *Server {
	return &Server{
		config:   cfg,
		exporter: exp,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	// setup listener
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// setup server
	var opts []grpc.ServerOption
	server := grpc.NewServer(opts...)
	pb.RegisterGitLabExporterServer(server, s.exporter)

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		slog.Info(fmt.Sprintf("Listening on %s", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		server.GracefulStop()
		return nil
	}
}
