package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
)

type Server struct {
	exporter *ClickHouseExporter
	health   *health.Server
}

func NewServer(exp *ClickHouseExporter) *Server {
	return &Server{
		exporter: exp,
		health:   health.NewServer(),
	}
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	// setup listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Listening on %s", listener.Addr().String()))

	// setup server
	var opts []grpc.ServerOption
	server := grpc.NewServer(opts...)

	healthgrpc.RegisterHealthServer(server, s.health)
	pb.RegisterGitLabExporterServer(server, s.exporter)

	s.health.SetServingStatus("", healthgrpc.HealthCheckResponse_UNKNOWN)

	// serve and monitor health
	g := new(errgroup.Group)
	g.Go(func() error {
		if err := s.getReady(ctx); err != nil {
			return err
		}

		return s.watchReadiness(ctx)
	})
	g.Go(func() error { return server.Serve(listener) })

	errChan := make(chan error)
	go func() {
		errChan <- g.Wait()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		s.health.Shutdown()
		server.GracefulStop()
		return nil
	}
}
