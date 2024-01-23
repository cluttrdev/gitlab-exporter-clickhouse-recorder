module github.com/cluttrdev/gitlab-clickhouse-exporter

go 1.21.4

replace github.com/cluttrdev/gitlab-exporter => ../gitlab-exporter

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.17.1
	github.com/cluttrdev/cli v0.0.0-20240117210602-b3e731e74746
	github.com/cluttrdev/gitlab-exporter v0.4.1
	github.com/google/go-cmp v0.5.9
	go.opentelemetry.io/proto/otlp v1.1.0
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/ClickHouse/ch-go v0.58.2 // indirect
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/creasty/defaults v1.7.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.6.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/paulmach/orb v0.10.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	go.opentelemetry.io/otel v1.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240116215550-a9fa1716bcac // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
