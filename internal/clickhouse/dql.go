package clickhouse

import "context"

func SelectPipelineMaxUpdatedAt(c *Client, ctx context.Context) (map[int64]float64, error) {
	const query string = `
        SELECT id, max(updated_at) AS updated_at
        FROM {db: Identifier}.pipelines
        GROUP BY id
        `
	var params = map[string]string{
		"db": c.dbName,
	}

	ctx = WithParameters(ctx, params)

	var results []struct {
		ID        int64   `ch:"id"`
		UpdatedAt float64 `ch:"updated_at"`
	}

	if err := c.Select(ctx, &results, query); err != nil {
		return nil, err
	}

	m := make(map[int64]float64, len(results))
	for _, res := range results {
		m[res.ID] = res.UpdatedAt
	}

	return m, nil
}

func SelectTableIDs[T int64 | string](c *Client, ctx context.Context, table string) (map[T]struct{}, error) {
	const query string = `
        SELECT id FROM {db: Identifier}.{table: Identifier}
        `
	var params = map[string]string{
		"db":    c.dbName,
		"table": table,
	}

	ctx = WithParameters(ctx, params)

	var results []struct {
		ID T `ch:"id"`
	}

	if err := c.Select(ctx, &results, query); err != nil {
		return nil, err
	}

	m := make(map[T]struct{}, len(results))
	for _, res := range results {
		m[res.ID] = struct{}{}
	}

	return m, nil
}

func SelectTraceSpanIDs(c *Client, ctx context.Context) (map[string]struct{}, error) {
	const query string = `
        SELECT id FROM {db: Identifier}.{table: Identifier}
        `
	var params = map[string]string{
		"db":    c.dbName,
		"table": "traces",
	}

	ctx = WithParameters(ctx, params)

	var results []struct {
		ID string `ch:"SpanId"`
	}

	if err := c.Select(ctx, &results, query); err != nil {
		return nil, err
	}

	m := make(map[string]struct{}, len(results))
	for _, res := range results {
		m[res.ID] = struct{}{}
	}

	return m, nil
}
