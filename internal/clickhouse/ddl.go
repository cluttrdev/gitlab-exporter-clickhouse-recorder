package clickhouse

import (
	"context"
	"fmt"
)

const (
	PipelinesTable          string = "pipelines"
	JobsTable               string = "jobs"
	SectionsTable           string = "sections"
	BridgesTable            string = "bridges"
	TestReportsTable        string = "testreports"
	TestSuitesTable         string = "testsuites"
	TestCasesTable          string = "testcases"
	LogEmbeddedMetricsTable string = "metrics"
	TraceSpansTable         string = "traces"
)

const (
	createPipelinesTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    id Int64,
    iid Int64,
    project_id Int64,
    status String,
    source String,
    ref String,
    sha String,
    before_sha String,
    tag Bool,
    yaml_errors String,
    created_at Float64,
    updated_at Float64,
    started_at Float64,
    finished_at Float64,
    committed_at Float64,
    duration Float64,
    queued_duration Float64,
    coverage Float64,
    web_url String
)
ENGINE ReplacingMergeTree(updated_at)
ORDER BY id
;
    `

	createJobsTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    coverage Float64,
    allow_failure Bool,
    created_at Float64,
    started_at Float64,
    finished_at Float64,
    erased_at Float64,
    duration Float64,
    queued_duration Float64,
    tag_list Array(String),
    id Int64,
    name String,
    pipeline Tuple(
        id Int64,
        project_id Int64,
        ref String,
        sha String,
        status String
    ),
    ref String,
    stage String,
    status String,
    failure_reason String,
    tag Bool,
    web_url String
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createBridgesTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    coverage Float64,
    allow_failure Bool,
    created_at Float64,
    started_at Float64,
    finished_at Float64,
    erased_at Float64,
    duration Float64,
    queued_duration Float64,
    id Int64,
    name String,
    pipeline Tuple(
        id Int64,
        iid Int64,
        project_id Int64,
        status String,
        source String,
        ref String,
        sha String,
        web_url String,
        created_at Float64,
        updated_at Float64
    ),
    ref String,
    stage String,
    status String,
    failure_reason String,
    tag Bool,
    web_url String,
    downstream_pipeline Tuple(
        id Int64,
        iid Int64,
        project_id Int64,
        status String,
        source String,
        ref String,
        sha String,
        web_url String,
        created_at Float64,
        updated_at Float64
    )
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createSectionsTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    id Int64,
    name String,
    job Tuple(
        id Int64,
        name String,
        status String
    ),
    pipeline Tuple(
        id Int64,
        project_id Int64,
        ref String,
        sha String,
        status String
    ),
    started_at Float64,
    finished_at Float64,
    duration Float64
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createTestReportsTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    id String,
    pipeline_id Int64,
    total_time Float64,
    total_count Int64,
    success_count Int64,
    failed_count Int64,
    skipped_count Int64,
    error_count Int64,
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createTestSuitesTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    id String,
    testreport_id String,
    pipeline_id Int64,
    name String,
    total_time Float64,
    total_count Int64,
    success_count Int64,
    failed_count Int64,
    skipped_count Int64,
    error_count Int64,
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createTestCasesTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    id String,
    testsuite_id String,
    testreport_id String,
    pipeline_id Int64,
    status String,
    name String,
    classname String,
    file String,
    execution_time Float64,
    system_output String,
    stack_trace String,
    attachment_url String,
    recent_failures Tuple(
        count Int64,
        base_branch String
    )
)
ENGINE ReplacingMergeTree()
ORDER BY id
;
    `

	createLogEmbeddedMetricsTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
    name String,
    labels Map(String, String),
    value Float64,
    timestamp Int64,
    job_id Int64,
    job_name String
)
ENGINE MergeTree()
ORDER BY (job_id, name, timestamp)
;
    `
)

const (
	// OpenTelemetry Traces
	// schemas taken from https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go

	createTracesTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
     Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
     TraceId String CODEC(ZSTD(1)),
     SpanId String CODEC(ZSTD(1)),
     ParentSpanId String CODEC(ZSTD(1)),
     TraceState String CODEC(ZSTD(1)),
     SpanName LowCardinality(String) CODEC(ZSTD(1)),
     SpanKind LowCardinality(String) CODEC(ZSTD(1)),
     ServiceName LowCardinality(String) CODEC(ZSTD(1)),
     ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     ScopeName String CODEC(ZSTD(1)),
     ScopeVersion String CODEC(ZSTD(1)),
     SpanAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     Duration Int64 CODEC(ZSTD(1)),
     StatusCode LowCardinality(String) CODEC(ZSTD(1)),
     StatusMessage String CODEC(ZSTD(1)),
     Events Nested (
         Timestamp DateTime64(9),
         Name LowCardinality(String),
         Attributes Map(LowCardinality(String), String)
     ) CODEC(ZSTD(1)),
     Links Nested (
         TraceId String,
         SpanId String,
         TraceState String,
         Attributes Map(LowCardinality(String), String)
     ) CODEC(ZSTD(1)),
     INDEX idx_trace_id TraceId TYPE bloom_filter(0.001) GRANULARITY 1,
     INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_span_attr_key mapKeys(SpanAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_span_attr_value mapValues(SpanAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_duration Duration TYPE minmax GRANULARITY 1
) ENGINE MergeTree()
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceName, SpanName, toUnixTimestamp(Timestamp), TraceId)
SETTINGS index_granularity=8192, ttl_only_drop_parts = 1
;
    `

	createTraceIdTsTableSQL = `
CREATE TABLE IF NOT EXISTS {db: Identifier}.{table: Identifier} (
     TraceId String CODEC(ZSTD(1)),
     Start DateTime64(9) CODEC(Delta, ZSTD(1)),
     End DateTime64(9) CODEC(Delta, ZSTD(1)),
     INDEX idx_trace_id TraceId TYPE bloom_filter(0.01) GRANULARITY 1
) ENGINE MergeTree()
ORDER BY (TraceId, toUnixTimestamp(Start))
SETTINGS index_granularity=8192
;
    `

	createTraceIdTsMaterializedViewSQL = `
CREATE MATERIALIZED VIEW IF NOT EXISTS {db: Identifier}.{view: Identifier}
TO {db: Identifier}.%s
AS SELECT
    TraceId,
    min(Timestamp) as Start,
    max(Timestamp) as End
FROM {db: Identifier}.%s
WHERE TraceId != ''
GROUP BY TraceId
;
    `

	createTraceViewSQL = `
CREATE VIEW IF NOT EXISTS ` + "`" + `%s` + "`" + `.%s AS
SELECT
    TraceId AS traceID,
    SpanId AS spanID,
    SpanName AS operationName,
    ParentSpanId AS parentSpanID,
    ServiceName AS serviceName,
    Duration / 1000000 AS duration,
    Timestamp AS startTime,
    arrayMap(key -> map('key', key, 'value', SpanAttributes[key]), mapKeys(SpanAttributes)) AS tags,
    arrayMap(key -> map('key', key, 'value', ResourceAttributes[key]), mapKeys(ResourceAttributes)) AS serviceTags
FROM ` + "`" + `%s` + "`" + `.%s
WHERE TraceId = {trace_id:String}
;
    `
)

func createTables(ctx context.Context, db string, client *Client) error {
	if err := createPipelinesTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create pipelines table: %w", err)
	}
	if err := createJobsTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create jobs table: %w", err)
	}
	if err := createSectionsTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create sections table: %w", err)
	}
	if err := createBridgesTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create bridges table: %w", err)
	}
	if err := createTestReportsTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create testreports table: %w", err)
	}
	if err := createTestSuitesTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create testsuites table: %w", err)
	}
	if err := createTestCasesTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create testcases table: %w", err)
	}
	if err := createLogEmbeddedMetricsTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create metrics table: %w", err)
	}

	if err := createTraceSpansTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create traces table: %w", err)
	}
	if err := createTraceIdTsTable(client, ctx, db); err != nil {
		return fmt.Errorf("exec create traceIdTs table: %w", err)
	}
	if err := createTraceIdTsMaterializedView(client, ctx, db); err != nil {
		return fmt.Errorf("exec create tracesIdTs materialized view: %w", err)
	}
	if err := createTraceView(client, ctx, db); err != nil {
		return fmt.Errorf("exec create trace view: %w", err)
	}

	return nil
}

func createPipelinesTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": PipelinesTable,
		}),
		createPipelinesTableSQL,
	)
}

func createJobsTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": JobsTable,
		}),
		createJobsTableSQL,
	)
}

func createSectionsTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": SectionsTable,
		}),
		createSectionsTableSQL,
	)
}

func createBridgesTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": BridgesTable,
		}),
		createBridgesTableSQL,
	)
}

func createTestReportsTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": TestReportsTable,
		}),
		createTestReportsTableSQL,
	)
}

func createTestSuitesTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": TestSuitesTable,
		}),
		createTestSuitesTableSQL,
	)
}

func createTestCasesTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": TestCasesTable,
		}),
		createTestCasesTableSQL,
	)
}

func createLogEmbeddedMetricsTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": LogEmbeddedMetricsTable,
		}),
		createLogEmbeddedMetricsTableSQL,
	)
}

func createTraceSpansTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": TraceSpansTable,
		}),
		createTracesTableSQL,
	)
}

func createTraceIdTsTable(c *Client, ctx context.Context, db string) error {
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":    db,
			"table": fmt.Sprintf("%s_trace_id_ts", TraceSpansTable),
		}),
		createTraceIdTsTableSQL,
	)
}

func createTraceIdTsMaterializedView(c *Client, ctx context.Context, db string) error {
	var query string = fmt.Sprintf(
		createTraceIdTsMaterializedViewSQL,
		fmt.Sprintf("%s_trace_id_ts", TraceSpansTable), // tableTo
		TraceSpansTable, // tableFrom
	)
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db":   db,
			"view": fmt.Sprintf("%s_trace_id_ts_mv", TraceSpansTable),
			// "tableTo": fmt.Sprintf("%s_trace_id_ts", TraceSpansTable),
			// "tableFrom": TraceSpansTable,
		}),
		query,
	)
}

func createTraceView(c *Client, ctx context.Context, db string) error {
	const viewName string = "trace_view"
	var query string = fmt.Sprintf(
		createTraceViewSQL,
		db, viewName, // {db: Identifier}.{view: Identifier}
		db, TraceSpansTable, // {db:Identifier}.{tableFrom: Identifier}
	)
	return c.Exec(
		WithParameters(ctx, map[string]string{
			"db": db,
			// "view": viewName,
			// "tableFrom": TraceSpansTable,
		}),
		query,
	)
}
