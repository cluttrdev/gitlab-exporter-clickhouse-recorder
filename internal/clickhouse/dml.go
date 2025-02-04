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
		err = batch.AppendStruct(&Pipeline{
			Id:        p.Id,
			Iid:       p.Iid,
			ProjectId: p.GetProject().GetId(),

			Name:          p.Name,
			Ref:           p.Ref,
			Sha:           p.Sha,
			Source:        p.Source,
			Status:        p.Status,
			FailureReason: p.FailureReason,

			CommittedAt: convertTimestamp(p.Timestamps.GetCommittedAt()),
			CreatedAt:   convertTimestamp(p.Timestamps.GetCreatedAt()),
			UpdatedAt:   convertTimestamp(p.Timestamps.GetUpdatedAt()),
			StartedAt:   convertTimestamp(p.Timestamps.GetStartedAt()),
			FinishedAt:  convertTimestamp(p.Timestamps.GetFinishedAt()),

			QueuedDuration: convertDuration(p.QueuedDuration),
			Duration:       convertDuration(p.Duration),

			Coverage: p.Coverage,

			Warnings:   p.Warnings,
			YamlErrors: p.YamlErrors,

			Child:                     p.Child,
			UpstreamPipelineId:        p.UpstreamPipeline.GetId(),
			UpstreamPipelineIid:       p.UpstreamPipeline.GetIid(),
			UpstreamPipelineProjectId: p.UpstreamPipeline.GetProject().GetId(),

			MergeRequestId:        p.MergeRequest.GetId(),
			MergeRequestIid:       p.MergeRequest.GetIid(),
			MergeRequestProjectId: p.MergeRequest.GetProject().GetId(),

			UserId: p.User.GetId(),
		})
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

		var jobKind string
		switch j.Kind {
		case typespb.JobKind_JOBKIND_UNSPECIFIED:
			jobKind = "unspecified"
		case typespb.JobKind_JOBKIND_BUILD:
			jobKind = "build"
		case typespb.JobKind_JOBKIND_BRIDGE:
			jobKind = "bridge"
		default:
			jobKind = "unknown"
		}

		err = batch.AppendStruct(&Job{
			Id:         j.Id,
			PipelineId: j.Pipeline.GetId(),
			ProjectId:  j.Pipeline.GetProject().GetId(),

			Name:          j.Name,
			Ref:           j.Ref,
			Status:        j.Status,
			FailureReason: j.FailureReason,

			CreatedAt:  convertTimestamp(j.Timestamps.GetCreatedAt()),
			QueuedAt:   convertTimestamp(j.Timestamps.GetQueuedAt()),
			StartedAt:  convertTimestamp(j.Timestamps.GetStartedAt()),
			FinishedAt: convertTimestamp(j.Timestamps.GetFinishedAt()),
			ErasedAt:   convertTimestamp(j.Timestamps.GetErasedAt()),

			QueuedDuration: convertDuration(j.QueuedDuration),
			Duration:       convertDuration(j.Duration),

			Coverage: j.Coverage,

			Stage:   j.Stage,
			TagList: j.Tags,

			AllowFailure: j.AllowFailure,
			Manual:       j.Manual,
			Retried:      j.Retried,
			Retryable:    j.Retryable,

			Kind:                        jobKind,
			DownstreamPipelineId:        j.DownstreamPipeline.GetId(),
			DownstreamPipelineIid:       j.DownstreamPipeline.GetIid(),
			DownstreamPipelineProjectId: j.DownstreamPipeline.GetProject().GetId(),

			RunnerId: j.Runner.GetId(),

			// deprecated
			Pipeline: []any{
				j.Pipeline.GetId(),
				j.Pipeline.GetProject().GetId(),
				j.Ref,
				"", // sha
				"", // status
			},
		})
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

func InsertBridges(c *Client, ctx context.Context, bridges []*typespb.Job) (int, error) {
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
			convertTimestamp(b.GetTimestamps().GetCreatedAt()),
			convertTimestamp(b.GetTimestamps().GetStartedAt()),
			convertTimestamp(b.GetTimestamps().GetFinishedAt()),
			convertTimestamp(b.GetTimestamps().GetErasedAt()),
			convertDuration(b.Duration),
			convertDuration(b.QueuedDuration),
			b.Id,
			b.Name,
			map[string]interface{}{
				"id":         b.GetPipeline().GetId(),
				"iid":        b.GetPipeline().GetIid(),
				"project_id": b.GetPipeline().GetProject().GetId(),
				"status":     "", // b.Pipeline.Status,
				"source":     "", // b.Pipeline.Source,
				"ref":        "", // b.Pipeline.Source,
				"sha":        "", // b.Pipeline.Sha,
				"web_url":    "", // b.Pipeline.WebUrl,
				"created_at": 0,  // convertTimestamp(b.Pipeline.CreatedAt),
				"updated_at": 0,  // convertTimestamp(b.Pipeline.UpdatedAt),
			},
			b.Ref,
			b.Stage,
			b.Status,
			b.FailureReason,
			false, // b.Tag,
			"",    // b.WebUrl,
			map[string]interface{}{
				"id":         b.DownstreamPipeline.GetId(),
				"iid":        b.DownstreamPipeline.GetIid(),
				"project_id": b.DownstreamPipeline.GetProject().GetId(),
				"status":     "", // b.DownstreamPipeline.Status,
				"source":     "", // b.DownstreamPipeline.Source,
				"ref":        "", // b.DownstreamPipeline.Source,
				"sha":        "", // b.DownstreamPipeline.Sha,
				"web_url":    "", // b.DownstreamPipeline.WebUrl,
				"created_at": 0,  // convertTimestamp(b.DownstreamPipeline.CreatedAt),
				"updated_at": 0,  // convertTimestamp(b.DownstreamPipeline.UpdatedAt),
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
		err = batch.AppendStruct(&Section{
			Id:         s.Id,
			JobId:      s.Job.GetId(),
			PipelineId: s.Job.GetPipeline().GetId(),
			ProjectId:  s.Job.GetPipeline().GetProject().GetId(),

			Name: s.Name,

			StartedAt:  convertTimestamp(s.StartedAt),
			FinishedAt: convertTimestamp(s.FinishedAt),

			Duration: convertDuration(s.Duration),

			// deprecated
			Job: []any{
				s.Job.GetId(),
				s.Job.GetName(),
				"", // status
			},
			Pipeline: []any{
				s.Job.GetPipeline().GetId(),
				s.Job.GetPipeline().GetProject().GetId(),
				"", // ref
				"", // sha
				"", // status
			},
		})
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
		err = batch.AppendStruct(&TestReport{
			Id:         tr.Id,
			PipelineId: tr.Pipeline.GetId(),
			ProjectId:  tr.Pipeline.GetProject().GetId(),

			TotalTime:    tr.TotalTime,
			TotalCount:   tr.TotalCount,
			ErrorCount:   tr.ErrorCount,
			FailedCount:  tr.FailedCount,
			SkippedCount: tr.SkippedCount,
			SuccessCount: tr.SuccessCount,
		})
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
		err = batch.AppendStruct(&TestSuite{
			Id:           ts.Id,
			TestReportId: ts.TestReport.GetId(),
			PipelineId:   ts.TestReport.GetPipeline().GetId(),
			ProjectId:    ts.TestReport.GetPipeline().GetProject().GetId(),

			Name:         ts.Name,
			TotalTime:    ts.TotalTime,
			TotalCount:   ts.TotalCount,
			ErrorCount:   ts.ErrorCount,
			FailedCount:  ts.FailedCount,
			SkippedCount: ts.SkippedCount,
			SuccessCount: ts.SuccessCount,

			Properties: convertTestProperties(ts.Properties),
		})
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
		err = batch.AppendStruct(&TestCase{
			Id:           tc.Id,
			TestSuiteId:  tc.TestSuite.GetId(),
			TestReportId: tc.TestSuite.GetTestReport().GetId(),
			PipelineId:   tc.TestSuite.GetTestReport().GetPipeline().GetId(),
			ProjectId:    tc.TestSuite.GetTestReport().GetPipeline().GetProject().GetId(),

			Status:        tc.Status,
			Name:          tc.Name,
			Classname:     tc.Classname,
			File:          tc.File,
			ExecutionTime: tc.ExecutionTime,
			SystemOutput:  tc.SystemOutput,
			AttachmentUrl: tc.AttachmentUrl,

			Properties: convertTestProperties(tc.Properties),
		})
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
		assignees_id := make([]int64, 0, len(mr.Participants.GetAssignees()))
		for _, a := range mr.Participants.GetAssignees() {
			assignees_id = append(assignees_id, a.Id)
		}
		reviewers_id := make([]int64, 0, len(mr.Participants.GetReviewers()))
		for _, a := range mr.Participants.GetReviewers() {
			reviewers_id = append(reviewers_id, a.Id)
		}
		approvers_id := make([]int64, 0, len(mr.Participants.GetApprovers()))
		for _, a := range mr.Participants.GetApprovers() {
			approvers_id = append(approvers_id, a.Id)
		}

		err = batch.AppendStruct(&MergeRequest{
			Id:        mr.Id,
			Iid:       mr.Iid,
			ProjectId: mr.Project.GetId(),

			CreatedAt: convertTimestamp(mr.Timestamps.GetCreatedAt()),
			UpdatedAt: convertTimestamp(mr.Timestamps.GetUpdatedAt()),
			MergedAt:  convertTimestamp(mr.Timestamps.GetMergedAt()),
			ClosedAt:  convertTimestamp(mr.Timestamps.GetClosedAt()),

			Name:   mr.Name,
			Title:  mr.Title,
			Labels: mr.Labels,

			State:       mr.State,
			MergeStatus: mr.MergeStatus,
			MergeError:  mr.MergeError,

			SourceProjectId: mr.SourceProjectId,
			SourceBranch:    mr.SourceBranch,
			TargetProjectId: mr.TargetProjectId,
			TargetBranch:    mr.TargetBranch,

			Additions:   mr.DiffStats.GetAdditions(),
			Changes:     mr.DiffStats.GetChanges(),
			Deletions:   mr.DiffStats.GetDeletions(),
			FileCount:   mr.DiffStats.GetFileCount(),
			CommitCount: mr.DiffStats.GetCommitCount(),

			BaseSha:         mr.DiffRefs.GetBaseSha(),
			HeadSha:         mr.DiffRefs.GetHeadSha(),
			StartSha:        mr.DiffRefs.GetStartSha(),
			MergeCommitSha:  mr.DiffRefs.GetMergeCommitSha(),
			RebaseCommitSha: mr.DiffRefs.GetRebaseCommitSha(),

			AuthorId:    mr.Participants.GetAuthor().GetId(),
			AssigneesId: assignees_id,
			ReviewersId: reviewers_id,
			ApproversId: approvers_id,
			MergeUserId: mr.Participants.GetMergeUser().GetId(),

			Approved:  mr.Flags.GetApproved(),
			Conflicts: mr.Flags.GetConflicts(),
			Draft:     mr.Flags.GetDraft(),
			Mergeable: mr.Flags.GetMergeable(),

			MilestoneId:        mr.Milestone.GetId(),
			MilestoneIid:       mr.Milestone.GetIid(),
			MilestoneProjectId: mr.Milestone.GetProject().GetId(),
		})
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
		err = batch.AppendStruct(&MergeRequestNoteEvent{
			Id:                    mre.Id,
			MergeRequestId:        mre.MergeRequest.GetId(),
			MergeRequestIid:       mre.MergeRequest.GetIid(),
			MergeRequestProjectId: mre.MergeRequest.GetProject().GetId(),

			CreatedAt:  convertTimestamp(mre.CreatedAt),
			UpdatedAt:  convertTimestamp(mre.UpdatedAt),
			ResolvedAt: convertTimestamp(mre.ResolvedAt),

			Type:     mre.Type,
			System:   mre.System,
			Internal: mre.Internal,

			AuthorId:   mre.AuthorId,
			Resolvable: mre.Resolveable,
			Resolved:   mre.Resolved,
			ResolverId: mre.ResolverId,
		})
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
		err = batch.AppendStruct(&Metric{
			Id:         string(m.Id),
			Iid:        m.Iid,
			JobId:      m.Job.GetId(),
			PipelineId: m.Job.GetPipeline().GetId(),
			ProjectId:  m.Job.GetPipeline().GetProject().GetId(),

			Name:      m.Name,
			Labels:    convertLabels(m.Labels),
			Value:     m.Value,
			Timestamp: m.Timestamp.AsTime().UnixMilli(),
		})
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
		err = batch.AppendStruct(&Project{
			Id:          p.Id,
			NamespaceId: p.Namespace.GetId(),

			Name:     p.Name,
			FullName: p.FullName,
			Path:     p.Path,
			FullPath: p.FullPath,

			Description: p.Description,
			Topics:      []string{},

			CreatedAt:      convertTimestamp(p.Timestamps.GetCreatedAt()),
			UpdatedAt:      convertTimestamp(p.Timestamps.GetUpdatedAt()),
			LastActivityAt: convertTimestamp(p.Timestamps.GetLastActivityAt()),

			JobArtifactsSize:      p.Statistics.GetJobArtifactsSize(),
			ContainerRegistrySize: p.Statistics.GetContainerRegistrySize(),
			LfsObjectsSize:        p.Statistics.GetLfsObjectsSize(),
			PackagesSize:          p.Statistics.GetPackagesSize(),
			PipelineArtifactsSize: p.Statistics.GetPipelineArtifactsSize(),
			RepositorySize:        p.Statistics.GetRepositorySize(),
			SnippetsSize:          p.Statistics.GetSnippetsSize(),
			StorageSize:           p.Statistics.GetStorageSize(),
			UploadsSize:           p.Statistics.GetUploadsSize(),
			WikiSize:              p.Statistics.GetWikiSize(),

			ForksCount:      p.Statistics.GetForksCount(),
			StarsCount:      p.Statistics.GetStarsCount(),
			CommitCount:     p.Statistics.GetCommitCount(),
			OpenIssuesCount: p.Statistics.GetOpenIssuesCount(),

			Archived:   p.Archived,
			Visibility: p.Visibility,
		})
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

	if spanCount == 0 {
		return 0, nil
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

func convertTestProperties(properties []*typespb.TestProperty) [][]string {
	ps := make([][]string, 0, len(properties))
	for _, p := range properties {
		ps = append(ps, []string{p.Name, p.Value})
	}
	return ps
}
