package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/retry"
)

func (s *Server) setServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	s.health.SetServingStatus(service, status)
}

func (s *Server) checkReadiness(ctx context.Context) error {
	return s.exporter.client.Ping(ctx)
}

func (s *Server) waitForReady(ctx context.Context) error {
	seconds := func(d time.Duration) time.Duration {
		s := math.Ceil(d.Seconds())
		return time.Duration(s) * time.Second
	}

	return retry.Do(
		func(ctx context.Context) error {
			err := s.checkReadiness(ctx)
			if err != nil {
				args := []any{
					"error", err,
				}

				v, ok := ctx.Value(retry.ContextValuesKey("retry")).(retry.ContextValues)
				if ok {
					args = append(args,
						"retry.attempt", v.Attempt,
						"retry.delay", fmt.Sprint(seconds(v.Delay)),
					)
				}

				slog.Error("Readiness check failed", args...)
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
			Jitter:       0.1, // +/- 10%
		}),
	)
}

func (s *Server) getReady(ctx context.Context) error {
	const service string = "" // i.e. system

	s.setServingStatus(service, healthpb.HealthCheckResponse_NOT_SERVING)

	if err := s.waitForReady(ctx); err != nil {
		return err
	}

	if err := s.exporter.client.CreateTables(ctx); err != nil {
		return err
	}

	if err := s.exporter.client.InitCache(ctx); err != nil {
		return err
	}

	s.setServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	return nil
}

func (s *Server) watchReadiness(ctx context.Context) error {
	const service string = "" // i.e. system

	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = s.checkReadiness(ctx)
			if err == nil { // everything okay
				break /* select */
			}

			// readiness check failed, waiting for it to succeed again
			s.setServingStatus(service, healthpb.HealthCheckResponse_NOT_SERVING)
			err = s.waitForReady(ctx)
			if err != nil {
				return err
			}
			slog.Info("Readiness check succeeded")
			s.setServingStatus(service, healthpb.HealthCheckResponse_SERVING)
		}

		time.Sleep(3 * time.Second)
	}
}
