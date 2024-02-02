package probes

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/probes/health"
	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/probes/pprof"
)

type HealthCheck = health.Check

type ServerConfig struct {
	Host string
	Port string

	LivenessCheck  HealthCheck
	ReadinessCheck HealthCheck

	Debug bool
}

type Server struct {
	cfg ServerConfig
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	// health check endpoints
	healthHandler := health.NewHandler()
	healthHandler.SetLivenessCheck(s.cfg.LivenessCheck)
	healthHandler.SetReadinessCheck(s.cfg.ReadinessCheck)

	mux.Handle("/health/", http.StripPrefix("/health", healthHandler))

	// debug endpoints
	if s.cfg.Debug {
		debugHandler := pprof.NewHandler()
		mux.Handle("/debug/pprof/", http.StripPrefix("/debug/pprof", debugHandler))
	}

	return mux
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := s.routes()

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port),
		Handler: mux,
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return srv.Shutdown(ctx)
	}
}
