package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"time"

	otlp_comonpb "go.opentelemetry.io/proto/otlp/common/v1"
	otlp_tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	pb "github.com/cluttrdev/gitlab-exporter/grpc/exporterpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertTimestamp(ts *timestamppb.Timestamp) float64 {
	return float64(ts.GetSeconds()) + float64(ts.GetNanos())*1.0e-09
}

func convertDuration(d *durationpb.Duration) float64 {
	return float64(d.GetSeconds()) + float64(d.GetNanos())*1.0e-09
}

func InsertPipelines(ctx context.Context, pipelines []*pb.Pipeline, c *Client) error {
	if c == nil {
		return errors.New("nil client")
	}

	const query string = `INSERT INTO {db: Identifier}.pipelines`
	var params = map[string]string{
		"db": c.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, p := range pipelines {
		err = batch.Append(
			p.Id,
			p.Iid,
			p.ProjectId,
			p.Status,
			p.Source,
			p.Ref,
			p.Sha,
			p.BeforeSha,
			p.Tag,
			p.YamlErrors,
			convertTimestamp(p.CreatedAt),
			convertTimestamp(p.UpdatedAt),
			convertTimestamp(p.StartedAt),
			convertTimestamp(p.FinishedAt),
			convertTimestamp(p.CommittedAt),
			convertDuration(p.Duration),
			convertDuration(p.QueuedDuration),
			p.Coverage,
			p.WebUrl,
		)
		if err != nil {
			return fmt.Errorf("error inserting pipelines: %w", err)
		}
	}

	return batch.Send()
}

func InsertJobs(ctx context.Context, jobs []*pb.Job, c *Client) error {
	const query string = `INSERT INTO {db: Identifier}.jobs`
	var params = map[string]string{
		"db": c.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, j := range jobs {
		err = batch.Append(
			j.Coverage,
			j.AllowFailure,
			convertTimestamp(j.CreatedAt),
			convertTimestamp(j.StartedAt),
			convertTimestamp(j.FinishedAt),
			convertTimestamp(j.ErasedAt),
			convertDuration(j.Duration),
			convertDuration(j.QueuedDuration),
			j.TagList,
			j.Id,
			j.Name,
			map[string]interface{}{
				"id":         j.Pipeline.Id,
				"project_id": j.Pipeline.ProjectId,
				"ref":        j.Pipeline.Ref,
				"sha":        j.Pipeline.Sha,
				"status":     j.Pipeline.Status,
			},
			j.Ref,
			j.Stage,
			j.Status,
			j.FailureReason,
			j.Tag,
			j.WebUrl,
		)
		if err != nil {
			return fmt.Errorf("error inserting jobs: %w", err)
		}
	}

	return batch.Send()
}

func InsertBridges(ctx context.Context, bridges []*pb.Bridge, c *Client) error {
	const query string = `INSERT INTO {db: Identifier}.bridges`
	var params = map[string]string{
		"db": c.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error inserting bridges: %w", err)
	}

	for _, b := range bridges {
		err = batch.Append(
			b.Coverage,
			b.AllowFailure,
			convertTimestamp(b.CreatedAt),
			convertTimestamp(b.StartedAt),
			convertTimestamp(b.FinishedAt),
			convertTimestamp(b.ErasedAt),
			convertDuration(b.Duration),
			convertDuration(b.QueuedDuration),
			b.Id,
			b.Name,
			map[string]interface{}{
				"id":         b.Pipeline.Id,
				"iid":        b.Pipeline.Iid,
				"project_id": b.Pipeline.ProjectId,
				"status":     b.Pipeline.Status,
				"source":     b.Pipeline.Source,
				"ref":        b.Pipeline.Source,
				"sha":        b.Pipeline.Sha,
				"web_url":    b.Pipeline.WebUrl,
				"created_at": convertTimestamp(b.Pipeline.CreatedAt),
				"updated_at": convertTimestamp(b.Pipeline.UpdatedAt),
			},
			b.Ref,
			b.Stage,
			b.Status,
			b.FailureReason,
			b.Tag,
			b.WebUrl,
			map[string]interface{}{
				"id":         b.DownstreamPipeline.Id,
				"iid":        b.DownstreamPipeline.Iid,
				"project_id": b.DownstreamPipeline.ProjectId,
				"status":     b.DownstreamPipeline.Status,
				"source":     b.DownstreamPipeline.Source,
				"ref":        b.DownstreamPipeline.Source,
				"sha":        b.DownstreamPipeline.Sha,
				"web_url":    b.DownstreamPipeline.WebUrl,
				"created_at": convertTimestamp(b.DownstreamPipeline.CreatedAt),
				"updated_at": convertTimestamp(b.DownstreamPipeline.UpdatedAt),
			},
		)
		if err != nil {
			return fmt.Errorf("[clickhouse.Client.InsertBridges] %w", err)
		}
	}

	return batch.Send()
}

func InsertSections(ctx context.Context, sections []*pb.Section, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.sections`
	var params = map[string]string{
		"db": client.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, s := range sections {
		err = batch.Append(
			s.Id,
			s.Name,
			map[string]interface{}{
				"id":     s.Job.Id,
				"name":   s.Job.Name,
				"status": s.Job.Status,
			},
			map[string]interface{}{
				"id":         s.Pipeline.Id,
				"project_id": s.Pipeline.ProjectId,
				"ref":        s.Pipeline.Ref,
				"sha":        s.Pipeline.Sha,
				"status":     s.Pipeline.Status,
			},
			convertTimestamp(s.StartedAt),
			convertTimestamp(s.FinishedAt),
			convertDuration(s.Duration),
		)
		if err != nil {
			return fmt.Errorf("error inserting sections: %w", err)
		}
	}

	return batch.Send()
}

func InsertTestReports(ctx context.Context, reports []*pb.TestReport, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.testreports`
	var params = map[string]string{
		"db": client.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, tr := range reports {
		var (
			ids    []int64
			names  []string
			times  []float64
			counts []int64
		)

		err = batch.Append(
			tr.Id,
			tr.PipelineId,
			tr.TotalTime,
			tr.TotalCount,
			tr.SuccessCount,
			tr.FailedCount,
			tr.SkippedCount,
			tr.ErrorCount,
			ids,
			names,
			times,
			counts,
		)
		if err != nil {
			return fmt.Errorf("error inserting testreports: %w", err)
		}
	}

	return batch.Send()
}

func InsertTestSuites(ctx context.Context, suites []*pb.TestSuite, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.testsuites`
	var params = map[string]string{
		"db": client.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, ts := range suites {
		var (
			ids      []int64
			statuses []string
			names    []string
		)

		err = batch.Append(
			ts.Id,
			map[string]interface{}{
				"id":          ts.TestreportId,
				"pipeline_id": ts.PipelineId,
			},
			ts.Name,
			ts.TotalTime,
			ts.TotalCount,
			ts.SuccessCount,
			ts.FailedCount,
			ts.SkippedCount,
			ts.ErrorCount,
			ids,
			statuses,
			names,
		)
		if err != nil {
			return fmt.Errorf("error inserting testsuites: %w", err)
		}
	}

	return batch.Send()
}

func InsertTestCases(ctx context.Context, cases []*pb.TestCase, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.testcases`
	var params = map[string]string{
		"db": client.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, tc := range cases {
		err = batch.Append(
			tc.Id,
			map[string]interface{}{
				"id": tc.TestsuiteId,
			},
			map[string]interface{}{
				"id":          tc.TestreportId,
				"pipeline_id": tc.PipelineId,
			},
			tc.Status,
			tc.Name,
			tc.Classname,
			tc.File,
			tc.ExecutionTime,
			tc.SystemOutput,
			tc.StackTrace,
			tc.AttachmentUrl,
			map[string]interface{}{
				"count":       tc.RecentFailures.Count,
				"base_branch": tc.RecentFailures.BaseBranch,
			},
		)
		if err != nil {
			return fmt.Errorf("[clickhouse.Client.InsertTestCases] %w", err)
		}
	}

	return batch.Send()
}

func InsertLogEmbeddedMetrics(ctx context.Context, metrics []*pb.LogEmbeddedMetric, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.{table: Identifier}`
	var params = map[string]string{
		"db":    client.dbName,
		"table": "log_embedded_metrics",
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, m := range metrics {
		err = batch.Append(
			m.Name,
			m.Labels,
			m.Value,
			convertTimestamp(m.Timestamp),
			m.Job.Id,
			m.Job.Name,
		)
		if err != nil {
			return fmt.Errorf("append batch:  %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}
	return nil
}

func InsertTraces(ctx context.Context, traces []*pb.Trace, client *Client) error {
	const query string = `INSERT INTO {db: Identifier}.traces`
	var params = map[string]string{
		"db": client.dbName,
	}

	ctx = WithParameters(ctx, params)

	batch, err := client.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing batch: %w", err)
	}

	for _, trace := range traces {
		for _, resourceSpans := range trace.Data.ResourceSpans {
			resourceAttrs := convertAttributes(resourceSpans.Resource.Attributes)
			serviceName := ""
			if sn, ok := resourceAttrs["service.name"]; ok {
				serviceName = sn
			}
			for _, scopeSpans := range resourceSpans.ScopeSpans {
				scopeName := scopeSpans.Scope.Name
				scopeVersion := scopeSpans.Scope.Version
				for _, span := range scopeSpans.Spans {
					spanAttrs := convertAttributes(span.Attributes)
					eventTimes, eventNames, eventAttrs := convertEvents(span.Events)
					linkTraceIDs, linkSpanIDs, linkStates, linkAttrs := convertLinks(span.Links)

					err = batch.Append(
						timeFromUnixNano(int64(span.StartTimeUnixNano)),
						span.TraceId,
						span.SpanId,
						span.ParentSpanId,
						span.TraceState,
						span.Name,
						span.Kind.String(),
						serviceName,
						resourceAttrs,
						scopeName,
						scopeVersion,
						spanAttrs,
						int64(span.EndTimeUnixNano-span.StartTimeUnixNano),
						span.Status.Code.String(),
						span.Status.Message,
						eventTimes,
						eventNames,
						eventAttrs,
						linkTraceIDs,
						linkSpanIDs,
						linkStates,
						linkAttrs,
					)

					if err != nil {
						return fmt.Errorf("error inserting traces: %w", err)
					}
				}
			}
		}
	}

	return batch.Send()
}

func timeFromUnixNano(ts int64) time.Time {
	const nsecPerSecond int64 = 1e09
	sec := ts / nsecPerSecond
	nsec := ts - (sec * nsecPerSecond)
	return time.Unix(sec, nsec)
}

func convertAttributes(list []*otlp_comonpb.KeyValue) map[string]string {
	attrs := make(map[string]string)

	for _, attr := range list {
		value, ok := attr.GetValue().Value.(*otlp_comonpb.AnyValue_StringValue)
		if ok {
			attrs[attr.Key] = value.StringValue
		}
	}

	return attrs
}

func convertEvents(events []*otlp_tracepb.Span_Event) ([]time.Time, []string, []map[string]string) {
	var (
		times []time.Time
		names []string
		attrs []map[string]string
	)
	for _, event := range events {
		times = append(times, timeFromUnixNano(int64(event.TimeUnixNano)))
		names = append(names, event.Name)
		attrs = append(attrs, convertAttributes(event.Attributes))
	}
	return times, names, attrs
}

func convertLinks(links []*otlp_tracepb.Span_Link) ([]string, []string, []string, []map[string]string) {
	var (
		traceIDs []string
		spanIDs  []string
		states   []string
		attrs    []map[string]string
	)
	for _, link := range links {
		traceIDs = append(traceIDs, string(link.TraceId))
		spanIDs = append(spanIDs, string(link.SpanId))
		states = append(states, link.TraceState)
		attrs = append(attrs, convertAttributes(link.Attributes))
	}
	return traceIDs, spanIDs, states, attrs
}
