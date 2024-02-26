package integration_tests

import (
	"context"
	"os"
	"testing"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	ch_tests "github.com/ClickHouse/clickhouse-go/v2/tests"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

const testSet string = "native"

func TestMain(m *testing.M) {
	env, err := ch_tests.CreateClickHouseTestEnvironment(testSet)
	if err != nil {
		panic(err)
	}
	defer env.Container.Terminate(context.Background())

	ch_tests.SetTestEnvironment(testSet, env)
	if err := ch_tests.CreateDatabase(testSet); err != nil {
		panic(err)
	}

	if err := CreateTables(testSet); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func GetTestClient(testSet string) (*clickhouse.Client, error) {
	te, err := ch_tests.GetTestEnvironment(testSet)
	if err != nil {
		return nil, err
	}

	opts := ch_tests.ClientOptionsFromEnv(te, ch.Settings{})
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
