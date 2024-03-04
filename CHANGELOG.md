# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.2] - 2024-03-04

### Changed

- Reduce memory required to cache entity ids

## [0.6.1] - 2024-02-26

### Changed

- Make ClickHouse client creation more flexible

### Fixed

- Catch jobs where pipeline reference is `nil`
- Improve table creation fucntions

## [0.6.0] - 2024-02-25

### Added

- gRPC server metrics

### Changed

- Switch to unary grpc calls

### Fixed

- Helm chart servie monitor template

## [0.5.3] - 2024-02-23

### Fixed

- Not waiting on retry ticker after stopped

## [0.5.2] - 2024-02-23

### Added

- HTTP probes server for metrics and debugging

### Fixed

- Some heap escape improvements

## [0.5.1] - 2024-02-22

### Fixed

- Adjust config env var prefix

## [0.5.0] - 2024-02-22

This release is due to renaming the project.

## [0.4.2] - 2024-02-20

## Fixed

- Trace spans insertion cache update
- Log-embedded metrics cache update

## [0.4.1] - 2024-02-19

### Fixed

- Rename RecordLogEmbeddedMetrics to RecordMetrics

## [0.4.0] - 2024-02-19

### Changed

- Adjust to changes in `gitlab-exporter v0.6.0`

## [0.3.1] - 2024-02-16

### Changed

- Check readiness every 3s

### Fixed

- Convert labels when inserting log embedded metrics
- Set initial serving status to UNKNOWN
- Log readiness check failures as errors

## [0.3.0] - 2024-02-14

### Changed

- Use gRPC health checks instead of HTTP probes

## [0.2.1] - 2024-02-12

### Changed

- Improve readiness and retry logic in run command

## [0.2.0] - 2024-02-02

### Added

- Structured logging
- Caching entity ids to avoid duplicates
- Deployment helm chart
- HTTP server for health and debug probes

### Changed

- Make recording methods more generic
- Return number of actually inserted entities

## [0.1.0] - 2024-01-26

Initial release.

<!-- Links -->
[Unreleased]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.6.2...HEAD
[0.6.2]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.6.1...v0.6.2
[0.6.1]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.5.3...v0.6.0
[0.5.3]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/cluttrdev/gitlab-exporter-clickhouse-recorder/releases/tag/v0.1.0
