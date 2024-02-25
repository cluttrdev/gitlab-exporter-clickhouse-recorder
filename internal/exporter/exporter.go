package exporter

import (
	"context"
	"log/slog"

	"github.com/cluttrdev/gitlab-exporter/protobuf/servicepb"
	"github.com/cluttrdev/gitlab-exporter/protobuf/typespb"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

type ClickHouseExporter struct {
	servicepb.UnimplementedGitLabExporterServer

	client *clickhouse.Client
}

func NewExporter(client *clickhouse.Client) *ClickHouseExporter {
	return &ClickHouseExporter{
		client: client,
	}
}

type insertFunc[T any] func(client *clickhouse.Client, ctx context.Context, data []*T) (int, error)

func record[T any](srv *ClickHouseExporter, ctx context.Context, data []*T, insert insertFunc[T]) (*servicepb.RecordSummary, error) {
	n, err := insert(srv.client, context.Background(), data)
	if err != nil {
		slog.Error("Failed to insert data", "error", err)
		return nil, err
	}

	return &servicepb.RecordSummary{
		RecordedCount: int32(n),
	}, nil
}

func (s *ClickHouseExporter) RecordPipelines(ctx context.Context, r *servicepb.RecordPipelinesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Pipeline](s, ctx, r.Data, clickhouse.InsertPipelines)
}

func (s *ClickHouseExporter) RecordJobs(ctx context.Context, r *servicepb.RecordJobsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Job](s, ctx, r.Data, clickhouse.InsertJobs)
}

func (s *ClickHouseExporter) RecordSections(ctx context.Context, r *servicepb.RecordSectionsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Section](s, ctx, r.Data, clickhouse.InsertSections)
}

func (s *ClickHouseExporter) RecordBridges(ctx context.Context, r *servicepb.RecordBridgesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Bridge](s, ctx, r.Data, clickhouse.InsertBridges)
}

func (s *ClickHouseExporter) RecordTestReports(ctx context.Context, r *servicepb.RecordTestReportsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestReport](s, ctx, r.Data, clickhouse.InsertTestReports)
}

func (s *ClickHouseExporter) RecordTestSuites(ctx context.Context, r *servicepb.RecordTestSuitesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestSuite](s, ctx, r.Data, clickhouse.InsertTestSuites)
}

func (s *ClickHouseExporter) RecordTestCases(ctx context.Context, r *servicepb.RecordTestCasesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestCase](s, ctx, r.Data, clickhouse.InsertTestCases)
}

func (s *ClickHouseExporter) RecordMetrics(ctx context.Context, r *servicepb.RecordMetricsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Metric](s, ctx, r.Data, clickhouse.InsertMetrics)
}

func (s *ClickHouseExporter) RecordTraces(ctx context.Context, r *servicepb.RecordTracesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Trace](s, ctx, r.Data, clickhouse.InsertTraces)
}
