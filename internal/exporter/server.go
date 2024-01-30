package exporter

import (
	"context"
	"io"
	"log/slog"

	"google.golang.org/grpc"

	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/clickhouse"
)

type ClickHouseServer struct {
	pb.UnimplementedGitLabExporterServer

	client *clickhouse.Client
}

func NewServer(client *clickhouse.Client) *ClickHouseServer {
	return &ClickHouseServer{
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

func record[T any](srv *ClickHouseServer, stream grpc.ServerStream, insert insertFunc[T]) error {
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

	return stream.SendMsg(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordPipelines(stream pb.GitLabExporter_RecordPipelinesServer) error {
	return record[pb.Pipeline](s, stream, clickhouse.InsertPipelines)
}

func (s *ClickHouseServer) RecordJobs(stream pb.GitLabExporter_RecordJobsServer) error {
	return record[pb.Job](s, stream, clickhouse.InsertJobs)
}

func (s *ClickHouseServer) RecordSections(stream pb.GitLabExporter_RecordSectionsServer) error {
	return record[pb.Section](s, stream, clickhouse.InsertSections)
}

func (s *ClickHouseServer) RecordBridges(stream pb.GitLabExporter_RecordBridgesServer) error {
	return record[pb.Bridge](s, stream, clickhouse.InsertBridges)
}

func (s *ClickHouseServer) RecordTestReports(stream pb.GitLabExporter_RecordTestReportsServer) error {
	return record[pb.TestReport](s, stream, clickhouse.InsertTestReports)
}

func (s *ClickHouseServer) RecordTestSuites(stream pb.GitLabExporter_RecordTestSuitesServer) error {
	return record[pb.TestSuite](s, stream, clickhouse.InsertTestSuites)
}

func (s *ClickHouseServer) RecordTestCases(stream pb.GitLabExporter_RecordTestCasesServer) error {
	return record[pb.TestCase](s, stream, clickhouse.InsertTestCases)
}

func (s *ClickHouseServer) RecordLogEmbeddedMetrics(stream pb.GitLabExporter_RecordLogEmbeddedMetricsServer) error {
	return record[pb.LogEmbeddedMetric](s, stream, clickhouse.InsertLogEmbeddedMetrics)
}

func (s *ClickHouseServer) RecordTraces(stream pb.GitLabExporter_RecordTracesServer) error {
	return record[pb.Trace](s, stream, clickhouse.InsertTraces)
}
