# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
[Unreleased]: https://github.com/cluttrdev/gitlab-clickhouse-exporter/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/cluttrdev/gitlab-clickhouse-exporter/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/cluttrdev/gitlab-clickhouse-exporter/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/cluttrdev/gitlab-clickhouse-exporter/releases/tag/v0.1.0
