package clickhouse

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type MigrationOptions struct {
	ClientConfig

	FileSystem fs.FS
	Path       string
}

var ErrMigrateNoChange = migrate.ErrNoChange

func MigrateUp(opts MigrationOptions) error {
	if opts.FileSystem == nil {
		return errors.New("missing migrations file system")
	}
	drv, err := iofs.New(opts.FileSystem, opts.Path)
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	q := url.Values{}
	q.Set("x-multi-statement", "true")
	q.Set("x-migrations-table", "schema_migrations")
	q.Set("x-migrations-table-engine", "MergeTree")
	// q.Set("x-cluster-name", "")

	dsn := url.URL{
		Scheme:   "clickhouse",
		Host:     fmt.Sprintf("%s:%s", opts.ClientConfig.Host, opts.ClientConfig.Port),
		Path:     opts.ClientConfig.Database,
		User:     url.UserPassword(opts.ClientConfig.User, opts.ClientConfig.Password),
		RawQuery: q.Encode(),
	}

	m, err := migrate.NewWithSourceInstance("iofs", drv, dsn.String())
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to apply up migrations: %w", err)
	}
	return nil
}
