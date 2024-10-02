package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	otlp_comonpb "go.opentelemetry.io/proto/otlp/common/v1"
	otlp_tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/cluttrdev/gitlab-exporter/protobuf/typespb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	PipelinesTable              string = "pipelines"
	JobsTable                   string = "jobs"
	SectionsTable               string = "sections"
	BridgesTable                string = "bridges"
	TestReportsTable            string = "testreports"
	TestSuitesTable             string = "testsuites"
	TestCasesTable              string = "testcases"
	MergeRequestsTable          string = "mergerequests"
	MergeRequestNoteEventsTable string = "mergerequest_noteevents"
	MetricsTable                string = "metrics"
	ProjectsTable               string = "projects"
	TraceSpansTable             string = "traces"
)

func convertTimestamp(ts *timestamppb.Timestamp) float64 {
	return float64(ts.GetSeconds()) + float64(ts.GetNanos())*1.0e-09
}

func convertDuration(d *durationpb.Duration) float64 {
	return float64(d.GetSeconds()) + float64(d.GetNanos())*1.0e-09
}

func InsertPipelines(c *Client, ctx context.Context, pipelines []*typespb.Pipeline) (int, error) {
	if c == nil {
		return 0, errors.New("nil client")
	}
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": PipelinesTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
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
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded pipelines", "received", len(pipelines), "inserted", n)

	return n, nil
}

func InsertJobs(c *Client, ctx context.Context, jobs []*typespb.Job) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": JobsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	var errs error
	for _, j := range jobs {
		if j.Pipeline == nil {
			errs = errors.Join(errs, fmt.Errorf("job without pipeline: %d", j.Id))
			continue
		}

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
			errs = errors.Join(errs, fmt.Errorf("append job %d to batch: %w", j.Id, err))
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded jobs", "received", len(jobs), "inserted", n)

	return n, errs
}

func InsertBridges(c *Client, ctx context.Context, bridges []*typespb.Bridge) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": BridgesTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
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
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded bridges", "received", len(bridges), "inserted", n)

	return n, nil
}

func InsertSections(c *Client, ctx context.Context, sections []*typespb.Section) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": SectionsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
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
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded sections", "received", len(sections), "inserted", n)

	return n, nil
}

func InsertTestReports(c *Client, ctx context.Context, reports []*typespb.TestReport) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": TestReportsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, tr := range reports {
		err = batch.Append(
			tr.Id,
			tr.PipelineId,
			tr.TotalTime,
			tr.TotalCount,
			tr.SuccessCount,
			tr.FailedCount,
			tr.SkippedCount,
			tr.ErrorCount,
		)
		if err != nil {
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded testreports", "received", len(reports), "inserted", n)

	return n, nil
}

func InsertTestSuites(c *Client, ctx context.Context, suites []*typespb.TestSuite) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": TestSuitesTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, ts := range suites {
		err = batch.Append(
			ts.Id,
			ts.TestreportId,
			ts.PipelineId,
			ts.Name,
			ts.TotalTime,
			ts.TotalCount,
			ts.SuccessCount,
			ts.FailedCount,
			ts.SkippedCount,
			ts.ErrorCount,
		)
		if err != nil {
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded testsuites", "received", len(suites), "inserted", n)

	return n, nil
}

func InsertTestCases(c *Client, ctx context.Context, cases []*typespb.TestCase) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": TestCasesTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, tc := range cases {
		err = batch.Append(
			tc.Id,
			tc.TestsuiteId,
			tc.TestreportId,
			tc.PipelineId,
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
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded testcases", "received", len(cases), "inserted", n)

	return n, nil
}

func InsertMergeRequests(c *Client, ctx context.Context, mrs []*typespb.MergeRequest) (int, error) {
	if c == nil {
		return 0, errors.New("nil client")
	}
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": MergeRequestsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, mr := range mrs {
		assignees_id := make([]int64, 0, len(mr.Assignees))
		for _, a := range mr.Assignees {
			assignees_id = append(assignees_id, a.Id)
		}
		reviewers_id := make([]int64, 0, len(mr.Reviewers))
		for _, a := range mr.Reviewers {
			reviewers_id = append(reviewers_id, a.Id)
		}

		err = batch.Append(
			mr.Id,
			mr.Iid,
			mr.ProjectId,
			convertTimestamp(mr.CreatedAt),
			convertTimestamp(mr.UpdatedAt),
			convertTimestamp(mr.MergedAt),
			convertTimestamp(mr.ClosedAt),
			mr.SourceProjectId,
			mr.TargetProjectId,
			mr.SourceBranch,
			mr.TargetBranch,
			mr.Title,
			mr.State,
			mr.DetailedMergeStatus,
			mr.Draft,
			mr.HasConflicts,
			mr.MergeError,
			mr.DiffRefs.BaseSha,
			mr.DiffRefs.HeadSha,
			mr.DiffRefs.StartSha,
			mr.GetAuthor().GetId(),
			mr.GetAssignee().GetId(),
			assignees_id,
			reviewers_id,
			mr.GetMergeUser().GetId(),
			mr.GetCloseUser().GetId(),
			mr.Labels,
			mr.Sha,
			mr.MergeCommitSha,
			mr.SquashCommitSha,
			mr.ChangesCount,
			mr.UserNotesCount,
			mr.Upvotes,
			mr.Downvotes,
			mr.GetPipeline().GetId(),
			mr.GetMilestone().GetId(),
			mr.WebUrl,
		)
		if err != nil {
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded mergerequests", "received", len(mrs), "inserted", n)

	return n, nil
}

func InsertMergeRequestNoteEvents(c *Client, ctx context.Context, mres []*typespb.MergeRequestNoteEvent) (int, error) {
	if c == nil {
		return 0, errors.New("nil client")
	}
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": MergeRequestNoteEventsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, mre := range mres {
		err = batch.Append(
			mre.Id,
			mre.MergerequestId,
			mre.MergerequestIid,
			mre.ProjectId,
			convertTimestamp(mre.CreatedAt),
			convertTimestamp(mre.UpdatedAt),
			mre.Type,
			mre.System,
			mre.AuthorId,
			mre.Resolveable,
			mre.Resolved,
			mre.ResolverId,
			mre.Confidential,
			mre.Internal,
		)
		if err != nil {
			return 0, fmt.Errorf("append batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded mergerequest_noteevents", "received", len(mres), "inserted", n)

	return n, nil
}

func InsertMetrics(c *Client, ctx context.Context, metrics []*typespb.Metric) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": MetricsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, m := range metrics {
		err = batch.Append(
			m.Id,
			m.Iid,
			m.JobId,
			m.Name,
			convertLabels(m.Labels),
			m.Value,
			convertTimestamp(m.Timestamp),
		)
		if err != nil {
			return 0, fmt.Errorf("append batch:  %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded metrics", "received", len(metrics), "inserted", n)

	return n, nil
}

func InsertProjects(c *Client, ctx context.Context, projects []*typespb.Project) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": ProjectsTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	for _, p := range projects {
		err = batch.Append(
			p.Id,
			p.GetNamespace().GetId(),
			p.GetOwner().GetId(),
			p.CreatorId,
			p.Name,
			p.NameWithNamespace,
			p.Path,
			p.PathWithNamespace,
			p.Description,
			p.Visibility,
			convertTimestamp(p.CreatedAt),
			convertTimestamp(p.LastActivityAt), //< real 'updated_at' not available, yet
			convertTimestamp(p.LastActivityAt),
			p.Topics,
			p.DefaultBranch,
			p.EmptyRepo,
			p.Archived,
			p.ForksCount,
			p.StarsCount,
			p.GetStatistics().GetCommitCount(),
			p.GetStatistics().GetStorageSize(),
			p.GetStatistics().GetRepositorySize(),
			p.GetStatistics().GetWikiSize(),
			p.GetStatistics().GetLfsObjectsSize(),
			p.GetStatistics().GetJobArtifactsSize(),
			p.GetStatistics().GetPipelineArtifactsSize(),
			p.GetStatistics().GetPackagesSize(),
			p.GetStatistics().GetSnippetsSize(),
			p.GetStatistics().GetUploadsSize(),
			p.OpenIssuesCount,
			p.WebUrl,
		)
		if err != nil {
			return 0, fmt.Errorf("append batch:  %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded projects", "received", len(projects), "inserted", n)

	return n, nil
}

func InsertTraces(c *Client, ctx context.Context, traces []*typespb.Trace) (int, error) {
	const query string = `INSERT INTO {db:Identifier}.{table:Identifier} SETTINGS async_insert=1`
	var params = map[string]string{
		"db":    c.dbName,
		"table": TraceSpansTable + "_in",
	}

	ctx = WithParameters(ctx, params)

	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}

	var spanCount int = 0
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
					spanCount++

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
						return 0, fmt.Errorf("append batch: %w", err)
					}
				}
			}
		}
	}

	if err := batch.Send(); err != nil {
		return -1, fmt.Errorf("send batch: %w", err)
	}

	n := batch.Rows()
	slog.Debug("Recorded trace spans", "received", spanCount, "inserted", n)

	return n, nil
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

func convertLabels(labels []*typespb.Metric_Label) map[string]string {
	m := make(map[string]string, len(labels))

	for _, l := range labels {
		m[l.Name] = l.Value
	}

	return m
}
