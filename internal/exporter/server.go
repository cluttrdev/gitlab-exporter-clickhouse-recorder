package exporter

import (
	"context"
	"io"

	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
	"google.golang.org/grpc"

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

func (s *ClickHouseServer) RecordPipelines(stream pb.GitLabExporter_RecordPipelinesServer) error {
	data, err := receive[pb.Pipeline](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertPipelines(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordJobs(stream pb.GitLabExporter_RecordJobsServer) error {
	data, err := receive[pb.Job](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertJobs(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordSections(stream pb.GitLabExporter_RecordSectionsServer) error {
	data, err := receive[pb.Section](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertSections(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordBridges(stream pb.GitLabExporter_RecordBridgesServer) error {
	data, err := receive[pb.Bridge](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertBridges(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordTestReports(stream pb.GitLabExporter_RecordTestReportsServer) error {
	data, err := receive[pb.TestReport](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertTestReports(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordTestSuites(stream pb.GitLabExporter_RecordTestSuitesServer) error {
	data, err := receive[pb.TestSuite](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertTestSuites(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordTestCases(stream pb.GitLabExporter_RecordTestCasesServer) error {
	data, err := receive[pb.TestCase](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertTestCases(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordLogEmbeddedMetrics(stream pb.GitLabExporter_RecordLogEmbeddedMetricsServer) error {
	data, err := receive[pb.LogEmbeddedMetric](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertLogEmbeddedMetrics(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}

func (s *ClickHouseServer) RecordTraces(stream pb.GitLabExporter_RecordTracesServer) error {
	data, err := receive[pb.Trace](stream)
	if err != io.EOF {
		return err
	}

	n, err := clickhouse.InsertTraces(context.Background(), data, s.client)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.RecordSummary{
		RecordedCount: int32(n),
	})
}
