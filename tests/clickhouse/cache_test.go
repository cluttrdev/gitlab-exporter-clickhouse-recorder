package clickhouse_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

func Test_CacheUpdatePipelines(t *testing.T) {
	cache := clickhouse.NewCache()

	data := map[int64]float64{
		1053344116: 1698520756.0,
		1053349645: 1698521748.0,
		1190130970: 1708897133.0,
	}

	updated := make(map[int64]bool, len(data))
	cache.UpdatePipelines(data, updated)

	expected := map[int64]bool{
		1053344116: true,
		1053349645: true,
		1190130970: true,
	}

	if diff := cmp.Diff(expected, updated); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}

	// ----

	data[1190130970] = 1709539234.0
	cache.UpdatePipelines(data, updated)

	expected = map[int64]bool{
		1053344116: false,
		1053349645: false,
		1190130970: true,
	}

	if diff := cmp.Diff(expected, updated); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func Test_CacheUpdateSections(t *testing.T) {
	cache := clickhouse.NewCache()

	data := map[int64]bool{
		6252785467: false,
		6252785469: false,
		6252785470: false,
		6252785472: false,
	}

	cache.UpdateSections(data)

	expected := map[int64]bool{
		6252785467: true,
		6252785469: true,
		6252785470: true,
		6252785472: true,
	}

	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}

	// ----

	data = map[int64]bool{
		6252785467: false,
		6252785469: false,
		6252785470: false,
		6252785472: false,
		6308490339: false,
	}

	cache.UpdateSections(data)

	expected = map[int64]bool{
		6252785467: false,
		6252785469: false,
		6252785470: false,
		6252785472: false,
		6308490339: true,
	}

	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}
