package clickhouse

import (
	"context"
	"fmt"
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
		Addr: []string{addr},
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
	}
}

func Connect(options *clickhouse.Options) (driver.Conn, error) {
	if options.Settings == nil {
		options.Settings = clickhouse.Settings{}
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
	pipelines, err := SelectPipelineMaxUpdatedAt(c, ctx)
	if err != nil {
		return err
	}
	c.cache.UpdatePipelines(pipelines, nil)

	jobs, err := SelectTableIDs[int64](c, ctx, JobsTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateJobs(keys(jobs), nil)

	sections, err := SelectTableIDs[int64](c, ctx, SectionsTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateSections(keys(sections), nil)

	bridges, err := SelectTableIDs[int64](c, ctx, BridgesTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateBridges(keys(bridges), nil)

	reports, err := SelectTableIDs[string](c, ctx, TestReportsTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestReports(keys(reports), nil)

	suites, err := SelectTableIDs[string](c, ctx, TestSuitesTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestSuites(keys(suites), nil)

	cases, err := SelectTableIDs[string](c, ctx, TestCasesTable, "id")
	if err != nil {
		return err
	}
	c.cache.UpdateTestCases(keys(cases), nil)

	metrics, err := SelectTableIDs[int64](c, ctx, LogEmbeddedMetricsTable, "job_id")
	if err != nil {
		return err
	}
	c.cache.UpdateLogEmbeddedMetrics(keys(metrics), nil)

	tracespans, err := SelectTraceSpanIDs(c, ctx)
	if err != nil {
		return err
	}
	c.cache.UpdateTraceSpans(keys(tracespans), nil)

	return nil
}

func keys[K int64 | string, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
