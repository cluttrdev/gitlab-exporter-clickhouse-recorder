#!/bin/bash

set -e

database=${INITDB_CLICKHOUSE_DATABASE:-'gitlab_ci'}
username=${INITDB_CLICKHOUSE_USER:-'default'}
password=${INITDB_CLICKHOUSE_PASSWORD:-''}

clickhouse client -n <<-EOSQL
    CREATE DATABASE IF NOT EXISTS ${database};

    CREATE USER IF NOT EXISTS ${username}
        IDENTIFIED WITH sha256_password BY '${password}'
        HOST ANY
        ;

    GRANT ALL ON ${database}.* TO ${username};
EOSQL
