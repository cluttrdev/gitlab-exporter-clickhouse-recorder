package integration_tests

import (
	"context"
	"os"
	"testing"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

const testSet string = "native"

func TestMain(m *testing.M) {
	env, err := CreateClickHouseTestEnvironment(testSet)
	if err != nil {
		panic(err)
	}
	defer env.Container.Terminate(context.Background())

	SetTestEnvironment(testSet, env)

	if err := CreateDatabase(testSet); err != nil {
		panic(err)
	}

	if err := CreateTables(testSet); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func GetTestClient(testSet string) (*clickhouse.Client, error) {
	te, err := GetTestEnvironment(testSet)
	if err != nil {
		return nil, err
	}

	opts := ClientOptionsFromEnv(te, ch.Settings{})
	opts.MaxOpenConns = 1

	conn, err := clickhouse.Connect(&opts)
	if err != nil {
		return nil, err
	}

	return clickhouse.NewClient(conn, te.Database), nil
}

func CreateTables(testSet string) error {
	client, err := GetTestClient(testSet)
	if err != nil {
		return err
	}

	return client.CreateTables(context.Background())
}
