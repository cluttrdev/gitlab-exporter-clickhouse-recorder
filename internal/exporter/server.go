package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cluttrdev/gitlab-exporter/protobuf/servicepb"
)

type Server struct {
	exporter *ClickHouseExporter
	health   *health.Server
	metrics  *grpcprom.ServerMetrics
}

func NewServer(exp *ClickHouseExporter) *Server {
	healthServer := health.NewServer()
	healthServer.SetServingStatus("" /* system */, healthgrpc.HealthCheckResponse_NOT_SERVING)

	metricsServer := grpcprom.NewServerMetrics()

	return &Server{
		exporter: exp,
		health:   healthServer,
		metrics:  metricsServer,
	}
}

func (s *Server) MetricsCollector() prometheus.Collector {
	return s.metrics
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	// setup grpc server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(s.metrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(s.metrics.StreamServerInterceptor()),
	)

	servicepb.RegisterGitLabExporterServer(grpcServer, s.exporter)
	healthgrpc.RegisterHealthServer(grpcServer, s.health)
	s.metrics.InitializeMetrics(grpcServer)

	// serve and monitor health
	g := &run.Group{}

	{ // serve grpc
		g.Add(func() error { // execute
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			slog.Info(fmt.Sprintf("Listening on %s", listener.Addr().String()))

			return grpcServer.Serve(listener)
		}, func(err error) { // interrupt
			s.health.Shutdown()
			grpcServer.GracefulStop()
			grpcServer.Stop()
		})
	}

	{ // monitor health
		ctx, cancel := context.WithCancel(ctx)
		g.Add(func() error { // execute
			if err := s.getReady(ctx); err != nil {
				return err
			}

			return s.watchReadiness(ctx)
		}, func(err error) { // interrupt
			cancel()
		})
	}

	{ // context handler
		ctx, cancel := context.WithCancel(ctx)
		g.Add(func() error { // execute
			<-ctx.Done()
			return ctx.Err()
		}, func(err error) { // interrupt
			cancel()
		})
	}

	return g.Run()
}
