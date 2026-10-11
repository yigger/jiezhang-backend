#!/bin/sh
set -eu

mysql_password=$(cat /run/dev-secrets/mysql-password)
export MYSQL_DSN="${MYSQL_USER}:${mysql_password}@tcp(${MYSQL_HOST}:3306)/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
export SESSION_TOKEN_SECRET="$(cat /run/dev-secrets/session-token-secret)"
exec "$@"
