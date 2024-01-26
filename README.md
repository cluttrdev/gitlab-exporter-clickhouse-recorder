# gitlab-clickhouse-exporter

`gitlab-clickhouse-exporter` serves a gRPC endpoint that records data from 
a [gitlab-exporter][github-gitlab-exporter] and exports it to a
[ClickHouse][clickhouse] database.

<p>
    <img src="./assets/project-overview.webp" />
    <img src="./assets/pipeline-trace.webp" />
</p>

## Installation

To install `gitlab-clickhouse-exporter` you can download a 
[prebuilt binary][prebuilt-binaries] that matches your system, e.g.

```shell
# download latest release archive
RELEASE_TAG=$(curl -sSfL https://api.github.com/repos/cluttrdev/gitlab-clickhouse-exporter/releases/latest | jq -r '.tag_name')
curl -sSfL -o /tmp/gitlab-clickhouse-exporter.tar.gz \
    https://github.com/cluttrdev/gitlab-clickhouse-exporter/releases/download/${RELEASE_TAG}/gitlab-clickhouse-exporter_${RELEASE_TAG}_linux_amd64.tar.gz
# extract executable binary into install dir (must exist)
INSTALL_DIR=$HOME/.local/bin
tar -C ${INSTALL_DIR} -zxof /tmp/gitlab-clickhouse-exporter.tar.gz gitlab-clickhouse-exporter
```

Alternatively, if you have the [Go][go-install] tools installed on your
machine, you can use

```shell
go install github.com/cluttrdev/gitlab-clickhouse-exporter@latest
```

## Usage

`gitlab-clickhouse-exporter` can either run in server mode or execute one-off
commands.

### Server Mode

To run `gitlab-clickhouse-exporter` in server mode use:

```shell
gitlab-clickhouse-exporter run --config CONFIG_FILE
```

This will start a gRPC server that exports recorded data to the configured 
ClickHouse database. See [Configuration](#configuration) for configuration options.

### Command Mode

`gitlab-clickhouse-exporter` supports commands that can be executed
individually. Use the following to get an overview of available commands:

```shell
gitlab-clickhouse-exporter -h
```

## Configuration

Configuration options can be specified in a config file that is passed to the
application using the `--config` command-line flag.

For an overview of available configuration options and their default values,
see [configs/gitlab-clickhouse-exporter.yaml](./configs/gitlab-clickhouse-exporter.yaml).

Common options can also be overridden with command-line flags and/or environment
variables, where flags take precedence.

| Flag                  | Environment Variable        | Default Value |
| ---                   | ---                         | ---           |
| --clickhouse-host     | `GLCHE_CLICKHOUSE_HOST`     | `"127.0.0.1"` |
| --clickhouse-port     | `GLCHE_CLICKHOUSE_PORT`     | `"9000"`      |
| --clickhouse-database | `GLCHE_CLICKHOUSE_DATABASE` | `"default"`   |
| --clickhouse-user     | `GLCHE_CLICKHOUSE_USER`     | `"default"`   |
| --clickhouse-password | `GLCHE_CLICKHOUSE_PASSWORD` | `""`          |
| --server-host         | `GLCHE_SERVER_HOST`         | `"0.0.0.0"`   |
| --server-port         | `GLCHE_SERVER_PORT`         | `"0"`         |

## Getting Started

To get up and running, have a look at the [quickstart](./examples/quickstart/README.md)
example which contains a `docker compose` setup to provision a ClickHouse server
and a Grafana instance that includes predefined dashboards.

## License

This project is licensed under the [MIT License](./LICENSE).

<!-- Links -->
[github-gitlab-exporter]: https://github.com/cluttrdev/gitlab-exporter
[clickhouse]: https://clickhouse.com/
[go-install]: https://go.dev/doc/install
[prebuilt-binaries]: https://github.com/cluttrdev/gitlab-clickhouse-exporter/releases/latest
