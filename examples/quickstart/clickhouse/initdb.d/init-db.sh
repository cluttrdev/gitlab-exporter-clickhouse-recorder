#!/bin/bash

set -e

database=${GLE_CLICKHOUSE_DATABASE:-'gitlab_ci'}
username=${GLE_CLICKHOUSE_USER:-'glche'}
password=${GLE_CLICKHOUSE_PASSWORD:-'glche'}

clickhouse client -n <<-EOSQL
    CREATE DATABASE IF NOT EXISTS ${database};

    CREATE USER IF NOT EXISTS ${username}
        IDENTIFIED WITH sha256_password BY '${password}'
        HOST ANY
        ;

    GRANT ALL ON ${database}.* TO ${username};
EOSQL
