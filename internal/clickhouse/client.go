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

func NewClient(cfg ClientConfig) (*Client, error) {
	client := &Client{
		cache: NewCache(),
	}

	if err := client.Configure(cfg); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Configure(cfg ClientConfig) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	conn, err := clickhouse.Open(&clickhouse.Options{
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
				{Name: "gitlab-clickhouse-exporter", Version: "v0.0.0+unknown"},
			},
		},
	})
	if err != nil {
		return err
	}

	c.Lock()
	c.conn = conn
	c.dbName = cfg.Database
	c.Unlock()
	return nil
}

func (c *Client) CheckReadiness(ctx context.Context) error {
	if err := c.conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return fmt.Errorf("clickhouse exception: [%d] %s", exception.Code, exception.Message)
		} else {
			return fmt.Errorf("error pinging clickhouse: %w", err)
		}
	}
	return nil
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
	c.cache.UpdatePipelines(pipelines)

	jobs, err := SelectTableIDs[int64](c, ctx, JobsTable)
	if err != nil {
		return err
	}
	c.cache.UpdateJobs(keys(jobs))

	sections, err := SelectTableIDs[int64](c, ctx, SectionsTable)
	if err != nil {
		return err
	}
	c.cache.UpdateSections(keys(sections))

	bridges, err := SelectTableIDs[int64](c, ctx, BridgesTable)
	if err != nil {
		return err
	}
	c.cache.UpdateBridges(keys(bridges))

	reports, err := SelectTableIDs[string](c, ctx, TestReportsTable)
	if err != nil {
		return err
	}
	c.cache.UpdateTestReports(keys(reports))

	suites, err := SelectTableIDs[string](c, ctx, TestSuitesTable)
	if err != nil {
		return err
	}
	c.cache.UpdateTestSuites(keys(suites))

	cases, err := SelectTableIDs[string](c, ctx, TestCasesTable)
	if err != nil {
		return err
	}
	c.cache.UpdateTestCases(keys(cases))

	metrics, err := SelectTableIDs[int64](c, ctx, LogEmbeddedMetricsTable)
	if err != nil {
		return err
	}
	c.cache.UpdateLogEmbeddedMetrics(keys(metrics))

	spans, err := SelectTraceSpanIDs(c, ctx)
	if err != nil {
		return err
	}
	c.cache.UpdateTraceSpans(keys(spans))

	return nil
}

func keys[K int64 | string, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
