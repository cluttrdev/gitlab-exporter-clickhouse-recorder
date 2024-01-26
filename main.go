package main

import (
	"fmt"
	"os"

	"github.com/cluttrdev/gitlab-clickhouse-exporter/internal/cmd"
)

var version string

func main() {
	cmd.Version = version

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
