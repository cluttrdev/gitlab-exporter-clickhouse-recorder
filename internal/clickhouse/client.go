package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Client struct {
	sync.RWMutex
	conn driver.Conn

	dbName string

	cache *Cache
}

type ClientConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func NewClient(conn driver.Conn, database string) *Client {
	return &Client{
		conn:   conn,
		cache:  NewCache(),
		dbName: database,
	}
}

func ClientOptions(cfg ClientConfig) clickhouse.Options {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "gitlab-exporter-clickhouse-recorder", Version: "v0.0.0+unknown"},
			},
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}
}

func Connect(options *clickhouse.Options) (driver.Conn, error) {
	if options.Settings == nil {
		options.Settings = clickhouse.Settings{
			"connect_timeout": 30,
		}
	}

	return clickhouse.Open(options)
}

func (c *Client) Ping(ctx context.Context) error {
	c.RLock()
	defer c.RUnlock()
	return c.conn.Ping(ctx)
}

func WithParameters(ctx context.Context, params map[string]string) context.Context {
	return clickhouse.Context(ctx, clickhouse.WithParameters(params))
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) error {
	c.RLock()
	defer c.RUnlock()
	return c.conn.Exec(ctx, query, args...)
}

func (c *Client) Select(ctx context.Context, dest any, query string, args ...any) error {
	c.RLock()
	defer c.RUnlock()
	return c.conn.Select(ctx, dest, query, args...)
}

func (c *Client) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	c.RLock()
	defer c.RUnlock()
	return c.conn.PrepareBatch(ctx, query)
}

func (c *Client) CreateTables(ctx context.Context) error {
	return createTables(ctx, c.dbName, c)
}

func (c *Client) InitCache(ctx context.Context) error {
	slog.Debug("Initializing pipelines cache...")
	pipelines, err := SelectPipelineMaxUpdatedAt(c, ctx)
	if err != nil {
		return err
	}
	c.cache.UpdatePipelines(pipelines, nil)
	slog.Debug("Initializing pipelines cache... done")

	slog.Debug("Initializing jobs cache...")
	jobs, err := SelectTableIDs[int64](c, ctx, JobsTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateJobs(keys(jobs), nil)
	slog.Debug("Initializing jobs cache... done")

	slog.Debug("Initializing sections cache...")
	sectionJobs, err := SelectTableIDs[int64](c, ctx, SectionsTable, "job.id")
	if err != nil {
		return err
	}
	c.cache.UpdateSections(keymap(sectionJobs))
	slog.Debug("Initializing sections cache... done")

	slog.Debug("Initializing bridges cache...")
	bridges, err := SelectTableIDs[int64](c, ctx, BridgesTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateBridges(keys(bridges), nil)
	slog.Debug("Initializing bridges cache... done")

	slog.Debug("Initializing testreports cache...")
	reports, err := SelectTableIDs[string](c, ctx, TestReportsTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestReports(keys(reports), nil)
	slog.Debug("Initializing testreports cache... done")

	slog.Debug("Initializing testsuites cache...")
	suites, err := SelectTableIDs[string](c, ctx, TestSuitesTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestSuites(keys(suites), nil)
	slog.Debug("Initializing testsuites cache... done")

	slog.Debug("Initializing testcases cache...")
	caseSuites, err := SelectTableIDs[string](c, ctx, TestCasesTable, "testsuite_id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestCases(keymap(caseSuites))
	slog.Debug("Initializing testcases cache... done")

	slog.Debug("Initializing metrics cache...")
	metrics, err := SelectTableIDs[int64](c, ctx, LogEmbeddedMetricsTable, "job_id")
	if err != nil {
		return err
	}
	c.cache.UpdateLogEmbeddedMetrics(keymap(metrics))
	slog.Debug("Initializing metrics cache... done")

	slog.Debug("Initializing tracespans cache...")
	tracespans, err := SelectTraceSpanIDs(c, ctx)
	if err != nil {
		return err
	}
	c.cache.UpdateTraceSpans(keys(tracespans), nil)
	slog.Debug("Initializing tracespans cache... done")

	return nil
}

func keys[K int64 | string, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

func keymap[K int64 | string, V any](m map[K]V) map[K]bool {
	km := make(map[K]bool, len(m))
	for k := range m {
		km[k] = false
	}
	return km
}
