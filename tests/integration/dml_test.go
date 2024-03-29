package integration_tests

import (
	"context"
	"testing"

	"github.com/cluttrdev/gitlab-exporter/protobuf/typespb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/internal/clickhouse"
)

func Test_InsertPipelines(t *testing.T) {
	client, err := GetTestClient(testSet)
	if err != nil {
		t.Error(err)
	}

	data := []*typespb.Pipeline{
		{
			Id: 1082136862, Iid: 12, ProjectId: 50817395,
			Status: "success", Source: "push", Ref: "main",
			Sha:            "e860ecdc74aee9aab22f6336a705bab05634c0c3",
			BeforeSha:      "7180ae586f19ae465387ddfa5f02522fa1521f6e",
			CreatedAt:      &timestamppb.Timestamp{Seconds: 1700690657, Nanos: 951000000},
			UpdatedAt:      &timestamppb.Timestamp{Seconds: 1700690933, Nanos: 886000000},
			StartedAt:      &timestamppb.Timestamp{Seconds: 1700690659, Nanos: 10000000},
			FinishedAt:     &timestamppb.Timestamp{Seconds: 1700690933, Nanos: 875000000},
			Duration:       &durationpb.Duration{Seconds: 273},
			QueuedDuration: &durationpb.Duration{Seconds: 1},
			WebUrl:         "https://gitlab.com/cluttrdev/gitlab-exporter/-/pipelines/1082136862",
		},
	}

	n, err := clickhouse.InsertPipelines(client, context.Background(), data)
	if err != nil {
		t.Error(err)
	}

	if n != len(data) {
		t.Errorf("Inserted %d pipelines, expected: %d", n, len(data))
	}
}

func Test_InsertJobs(t *testing.T) {
	client, err := GetTestClient(testSet)
	if err != nil {
		t.Error(err)
	}

	data := []*typespb.Job{
		{
			Pipeline: &typespb.PipelineReference{
				Id:        1082136862,
				ProjectId: 50817395,
				Ref:       "main",
				Sha:       "e860ecdc74aee9aab22f6336a705bab05634c0c3",
				Status:    "success",
			},
			Id: 5599404160, Name: "test", Ref: "main", Stage: "test", Status: "success",
			CreatedAt:      &timestamppb.Timestamp{Seconds: 1700690657, Nanos: 999000000},
			StartedAt:      &timestamppb.Timestamp{Seconds: 1700690851, Nanos: 366000000},
			FinishedAt:     &timestamppb.Timestamp{Seconds: 1700690933, Nanos: 765000000},
			Duration:       &durationpb.Duration{Seconds: 82, Nanos: 399463000},
			QueuedDuration: &durationpb.Duration{Nanos: 359749000},
			WebUrl:         "https://gitlab.com/cluttrdev/gitlab-exporter/-/jobs/5599404160",
		},
		{
			Pipeline: nil,
			Id:       42,
		},
	}

	n, err := clickhouse.InsertJobs(client, context.Background(), data)
	if err == nil {
		t.Errorf("Expected error due to job without pipeline, got `nil`")
	} else if err.Error() != "job without pipeline: 42" {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(data)-1 {
		t.Errorf("Inserted %d jobs, expected: %d", n, len(data)-1)
	}
}

func Test_InsertTestCases(t *testing.T) {
	client, err := GetTestClient(testSet)
	if err != nil {
		t.Error(err)
	}

	data := []*typespb.TestCase{
		{Id: "6252785472-1", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-2", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-3", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-4", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-5", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-6", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-7", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-8", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-9", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6252785472-10", TestsuiteId: "6252785472", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
	}

	n, err := clickhouse.InsertTestCases(client, context.Background(), data)
	if err != nil {
		t.Error(err)
	}

	if n != 10 {
		t.Errorf("Inserted %d testcases, expected: %d", n, 10)
	}

	// ----

	data = append(data, []*typespb.TestCase{
		{Id: "6308490339-1", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-2", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-3", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-4", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-5", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-6", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-7", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-8", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-9", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
		{Id: "6308490339-10", TestsuiteId: "6308490339", TestreportId: "1190130970", PipelineId: 1190130970, RecentFailures: &typespb.TestCase_RecentFailures{}},
	}...,
	)

	n, err = clickhouse.InsertTestCases(client, context.Background(), data)
	if err != nil {
		t.Error(err)
	}

	if n != 10 {
		t.Errorf("Inserted %d testcases, expected: %d", n, 10)
	}
}
