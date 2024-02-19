package exporter

import (
	"context"
	"io"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/cluttrdev/gitlab-exporter/protobuf/servicepb"
	"github.com/cluttrdev/gitlab-exporter/protobuf/typespb"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
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

func receive[T any](stream grpc.ServerStream) ([]*T, error) {
	var data []*T
	for {
		msg := new(T)
		if err := stream.RecvMsg(msg); err != nil {
			return data, err
		}
		data = append(data, msg)
	}
}

type insertFunc[T any] func(client *clickhouse.Client, ctx context.Context, data []*T) (int, error)

func record[T any](srv *ClickHouseExporter, stream grpc.ServerStream, insert insertFunc[T]) error {
	data, err := receive[T](stream)
	if err != io.EOF {
		slog.Error("Failed to receive data", "error", err)
		return err
	}

	n, err := insert(srv.client, context.Background(), data)
	if err != nil {
		slog.Error("Failed to insert data", "error", err)
		return err
	}

	return stream.SendMsg(&servicepb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseExporter) RecordPipelines(stream servicepb.GitLabExporter_RecordPipelinesServer) error {
	return record[typespb.Pipeline](s, stream, clickhouse.InsertPipelines)
}

func (s *ClickHouseExporter) RecordJobs(stream servicepb.GitLabExporter_RecordJobsServer) error {
	return record[typespb.Job](s, stream, clickhouse.InsertJobs)
}

func (s *ClickHouseExporter) RecordSections(stream servicepb.GitLabExporter_RecordSectionsServer) error {
	return record[typespb.Section](s, stream, clickhouse.InsertSections)
}

func (s *ClickHouseExporter) RecordBridges(stream servicepb.GitLabExporter_RecordBridgesServer) error {
	return record[typespb.Bridge](s, stream, clickhouse.InsertBridges)
}

func (s *ClickHouseExporter) RecordTestReports(stream servicepb.GitLabExporter_RecordTestReportsServer) error {
	return record[typespb.TestReport](s, stream, clickhouse.InsertTestReports)
}

func (s *ClickHouseExporter) RecordTestSuites(stream servicepb.GitLabExporter_RecordTestSuitesServer) error {
	return record[typespb.TestSuite](s, stream, clickhouse.InsertTestSuites)
}

func (s *ClickHouseExporter) RecordTestCases(stream servicepb.GitLabExporter_RecordTestCasesServer) error {
	return record[typespb.TestCase](s, stream, clickhouse.InsertTestCases)
}

func (s *ClickHouseExporter) RecordLogEmbeddedMetrics(stream servicepb.GitLabExporter_RecordMetricsServer) error {
	return record[typespb.Metric](s, stream, clickhouse.InsertLogEmbeddedMetrics)
}

func (s *ClickHouseExporter) RecordTraces(stream servicepb.GitLabExporter_RecordTracesServer) error {
	return record[typespb.Trace](s, stream, clickhouse.InsertTraces)
}
