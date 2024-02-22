# Quickstart Example

To get started, run:
```shell
docker compose up -d
```

The ClickHouse server will listen on `127.0.0.1:9000` and have the following
database and user credentials created:
```
database: gitlab_ci
user:     glche
password: glche
```
See the
[config.xml](./clickhouse/config.xml),
[users.xml](./clickhouse/users.xml) and
[init-db.sh](./clickhouse/initdb.d/init-db.sh)
files for more details.

Then, create a configuration file or set the necessary environment variables
and run `gitlab-exporter-clickhouse-recorder`.
```shell
# create simple config file with the database settings
cat <<EOF > config.yaml
clickhouse:
  database: "gitlab_ci"
  user: "glche"
  password: "glche"
EOF

# and/or set environment variables
export GLCHR_SERVER_HOST=127.0.0.1
export GLCHR_SERVER_PORT=36275

# run the server
gitlab-exporter-clickhouse-recorder run --config config.yaml
```

You should now be able to login to Grafana at <http://localhost:3000> (using
the default `admin:admin` credentials) and explore the data 
(once you start sending data using [gitlab-exporter][gh-gitlab-exporter].

<!-- Links -->
[gh-gitlab-exporter]: https://github.com/cluttrdev/gitlab-exporter

