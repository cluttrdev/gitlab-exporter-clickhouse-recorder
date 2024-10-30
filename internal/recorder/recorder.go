package recorder

import (
	"context"
	"log/slog"

	"github.com/cluttrdev/gitlab-exporter/protobuf/servicepb"
	"github.com/cluttrdev/gitlab-exporter/protobuf/typespb"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

type ClickHouseRecorder struct {
	servicepb.UnimplementedGitLabExporterServer

	client *clickhouse.Client
}

func New(client *clickhouse.Client) *ClickHouseRecorder {
	return &ClickHouseRecorder{
		client: client,
	}
}

type insertFunc[T any] func(client *clickhouse.Client, ctx context.Context, data []*T) (int, error)

func record[T any](srv *ClickHouseRecorder, ctx context.Context, data []*T, insert insertFunc[T]) (*servicepb.RecordSummary, error) {
	if len(data) == 0 {
		return &servicepb.RecordSummary{}, nil
	}

	n, err := insert(srv.client, context.Background(), data)
	if err != nil {
		slog.Error("Failed to insert data", "error", err)
		return nil, err
	}

	return &servicepb.RecordSummary{
		RecordedCount: int32(n),
	}, nil
}

func (s *ClickHouseRecorder) RecordPipelines(ctx context.Context, r *servicepb.RecordPipelinesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Pipeline](s, ctx, r.Data, clickhouse.InsertPipelines)
}

func (s *ClickHouseRecorder) RecordJobs(ctx context.Context, r *servicepb.RecordJobsRequest) (*servicepb.RecordSummary, error) {
	var (
		builds  []*typespb.Job
		bridges []*typespb.Job
	)
	for _, job := range r.Data {
		if job.Kind == typespb.JobKind_JOBKIND_BRIDGE {
			bridges = append(bridges, job)
		} else {
			builds = append(builds, job)
		}
	}

	buildsSummary, err := record[typespb.Job](s, ctx, builds, clickhouse.InsertJobs)
	if err != nil {
		return buildsSummary, err
	}
	bridgesSummary, err := record[typespb.Job](s, ctx, bridges, clickhouse.InsertBridges)
	if err != nil {
		return bridgesSummary, err
	}

	return &servicepb.RecordSummary{
		RecordedCount: buildsSummary.RecordedCount + bridgesSummary.RecordedCount,
	}, nil
}

func (s *ClickHouseRecorder) RecordSections(ctx context.Context, r *servicepb.RecordSectionsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Section](s, ctx, r.Data, clickhouse.InsertSections)
}

func (s *ClickHouseRecorder) RecordTestReports(ctx context.Context, r *servicepb.RecordTestReportsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestReport](s, ctx, r.Data, clickhouse.InsertTestReports)
}

func (s *ClickHouseRecorder) RecordTestSuites(ctx context.Context, r *servicepb.RecordTestSuitesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestSuite](s, ctx, r.Data, clickhouse.InsertTestSuites)
}

func (s *ClickHouseRecorder) RecordTestCases(ctx context.Context, r *servicepb.RecordTestCasesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.TestCase](s, ctx, r.Data, clickhouse.InsertTestCases)
}

func (s *ClickHouseRecorder) RecordMergeRequests(ctx context.Context, r *servicepb.RecordMergeRequestsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.MergeRequest](s, ctx, r.Data, clickhouse.InsertMergeRequests)
}

func (s *ClickHouseRecorder) RecordMergeRequestNoteEvents(ctx context.Context, r *servicepb.RecordMergeRequestNoteEventsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.MergeRequestNoteEvent](s, ctx, r.Data, clickhouse.InsertMergeRequestNoteEvents)
}

func (s *ClickHouseRecorder) RecordProjects(ctx context.Context, r *servicepb.RecordProjectsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Project](s, ctx, r.Data, clickhouse.InsertProjects)
}

func (s *ClickHouseRecorder) RecordMetrics(ctx context.Context, r *servicepb.RecordMetricsRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Metric](s, ctx, r.Data, clickhouse.InsertMetrics)
}

func (s *ClickHouseRecorder) RecordTraces(ctx context.Context, r *servicepb.RecordTracesRequest) (*servicepb.RecordSummary, error) {
	return record[typespb.Trace](s, ctx, r.Data, clickhouse.InsertTraces)
}
